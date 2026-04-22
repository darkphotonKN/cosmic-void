package member

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/auth"
	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/models"
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commonoutbox "github.com/darkphotonKN/cosmic-void-server/common/outbox"
	"github.com/darkphotonKN/cosmic-void-server/common/utils/cache"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/codes"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service struct {
	db              *sqlx.DB
	Repo            Repository
	publishCh       commonbroker.Publisher
	cache           cache.Cache
	outboxPublisher commonoutbox.OutboxPublisher
}

type Repository interface {
	Create(ctx context.Context, name, email, password string) (uuid.UUID, error)
	CreateTx(ctx context.Context, tx *sqlx.Tx, name, email, password string) (uuid.UUID, error)
	UpdatePassword(ctx context.Context, params MemberUpdatePasswordParams) error
	UpdateMemberInfo(ctx context.Context, id uuid.UUID, name, status string) error
	GetByIdWithPassword(ctx context.Context, id uuid.UUID) (*models.Member, error)
	GetById(ctx context.Context, id uuid.UUID) (*models.Member, error)
	GetMemberByEmail(ctx context.Context, email string) (*models.Member, error)
	CreateDefaultMembers(ctx context.Context, members []CreateDefaultMember) error
	UpdateAvatarURL(ctx context.Context, memberID uuid.UUID, avatarURL string) (*models.Member, error)
	UpdateAvatarURLTx(ctx context.Context, tx *sqlx.Tx, memberID uuid.UUID, avatarURL string) (*models.Member, error)
	GetStripeCustomerID(ctx context.Context, memberID uuid.UUID) (string, error)
	SetStripeCustomerID(ctx context.Context, memberID uuid.UUID, customerID string) error
	UpdateSubscriptionStatus(ctx context.Context, memberID uuid.UUID, productID, status string) error
}

func NewService(db *sqlx.DB, repo Repository, publishCh commonbroker.Publisher, cacheService cache.Cache, outboxPublisher commonoutbox.OutboxPublisher) *service {
	return &service{
		db:              db,
		Repo:            repo,
		publishCh:       publishCh,
		cache:           cacheService,
		outboxPublisher: outboxPublisher,
	}
}

func memberToProto(m *models.Member) *pb.Member {
	if m == nil {
		return nil
	}

	protoMember := &pb.Member{
		Id:            m.ID.String(),
		Name:          m.Name,
		Email:         m.Email,
		Status:        int32(stringToInt(m.Status)),
		AverageRating: float32(m.AverageRating),
		CreatedAt:     timestamppb.New(m.CreatedAt),
		UpdatedAt:     timestamppb.New(m.UpdatedAt),
		Role:          m.Role,
	}

	// Include avatar_url if it exists
	if m.AvatarURL != nil {
		protoMember.AvatarUrl = *m.AvatarURL
	}

	return protoMember
}

func stringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func (s *service) GetMember(ctx context.Context, req *pb.GetMemberRequest) (*pb.Member, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	member, err := s.Repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	return memberToProto(member), nil
}

func (s *service) CreateMember(ctx context.Context, req *pb.CreateMemberRequest) (*pb.Member, error) {

	// span for entire function
	ctx, span := serviceTracer.Start(ctx, "service.CreateMember")
	defer span.End()

	// span just for password
	ctx, hashSpan := serviceTracer.Start(ctx, "service.HashPassword")
	// hash the password
	hashedPw, err := s.HashPassword(req.Password)
	hashSpan.End()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "password hashing failed")
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Build the outbox payload up-front so a marshal failure aborts before
	// we open a transaction.
	payload := commonconstants.MemberSignedUpEventPayload{
		EventID:    uuid.NewString(),
		UserID:     "", // filled in after member row is created
		Name:       req.Name,
		Email:      req.Email,
		SignedUpAt: "", // TODO: update this to legit date
	}

	// Atomic write: member row + outbox row share a single transaction so
	// either both are visible or neither is. This is the core guarantee of
	// the transactional outbox pattern — without it, an RPC can succeed
	// while the event is lost (or vice versa).
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	// Safety net: if we return before Commit, roll back. Commit below
	// makes this a no-op.
	defer func() {
		_ = tx.Rollback()
	}()

	memberId, err := s.Repo.CreateTx(ctx, tx, req.Name, req.Email, hashedPw)
	if err != nil {
		return nil, err
	}

	payload.UserID = memberId.String()
	marshalledPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if err = s.outboxPublisher.CreateOutboxTx(ctx, tx, commonoutbox.OutboxParams{
		RoutingKey: commonconstants.MemberSignedUpEvent,
		Exchange:   commonconstants.AuthEventsExchange,
		Payload:    marshalledPayload,
	}); err != nil {
		return nil, fmt.Errorf("write outbox: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit member+outbox: %w", err)
	}

	// Post-commit read: member is guaranteed to exist. A failure here
	// does not affect the event — the outbox row is already durable and
	// the worker will publish it.
	member, err := s.Repo.GetById(ctx, memberId)
	if err != nil {
		return nil, err
	}

	return memberToProto(member), nil
}

func (s *service) LoginMember(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	member, err := s.Repo.GetMemberByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("could not find member with provided email: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(member.Password), []byte(req.Password)); err != nil {
		return nil, commonconstants.ErrUnauthorized
	}

	// generate tokens
	accessExpiryTime := time.Minute * 60
	refreshExpiryTime := time.Hour * 24 * 7

	accessToken, err := auth.GenerateJWT(*member, commonconstants.Access, accessExpiryTime)
	if err != nil {
		return nil, fmt.Errorf("error generating access token: %w", err)
	}

	refreshToken, err := auth.GenerateJWT(*member, commonconstants.Refresh, refreshExpiryTime)
	if err != nil {
		return nil, fmt.Errorf("error generating refresh token: %w", err)
	}

	slog.Debug("Generated tokens", "accessToken", accessToken)

	return &pb.LoginResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresIn:  int32(accessExpiryTime.Seconds()),
		RefreshExpiresIn: int32(refreshExpiryTime.Seconds()),
		MemberInfo:       memberToProto(member),
	}, nil
}

// UpdateMemberInfo implements the gRPC UpdateMemberInfo method
func (s *service) UpdateMemberInfo(ctx context.Context, req *pb.UpdateMemberInfoRequest) (*pb.Member, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	// Update member info
	err = s.Repo.UpdateMemberInfo(ctx, id, req.Name, req.Status)
	if err != nil {
		return nil, err
	}

	// Get the updated member
	member, err := s.Repo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	return memberToProto(member), nil
}

func (s *service) UpdateMemberPassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*pb.UpdatePasswordResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	// Get the member with password
	member, err := s.Repo.GetByIdWithPassword(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if new passwords match
	if req.NewPassword != req.RepeatNewPassword {
		return &pb.UpdatePasswordResponse{
			Success: false,
			Message: "New passwords do not match",
		}, errors.New("new passwords do not match")
	}

	// Verify current password
	isSame, err := s.ComparePasswords(member.Password, req.CurrentPassword)
	if !isSame || err != nil {
		return &pb.UpdatePasswordResponse{
			Success: false,
			Message: "Current password is incorrect",
		}, errors.New("current password is incorrect")
	}

	// Hash the new password
	hashedPw, err := s.HashPassword(req.NewPassword)
	if err != nil {
		return &pb.UpdatePasswordResponse{
			Success: false,
			Message: "Error hashing password",
		}, fmt.Errorf("error hashing password: %w", err)
	}

	// Update the password in the database
	params := MemberUpdatePasswordParams{
		ID:       id,
		Password: hashedPw,
	}

	err = s.Repo.UpdatePassword(ctx, params)
	if err != nil {
		return &pb.UpdatePasswordResponse{
			Success: false,
			Message: "Error updating password",
		}, err
	}

	return &pb.UpdatePasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}, nil
}

func (s *service) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// validate the token using auth package
	claims, err := auth.ValidateJWT(req.Token)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid:    false,
			MemberId: "",
		}, err
	}

	return &pb.ValidateTokenResponse{
		Valid:    true,
		MemberId: claims.ID,
	}, nil
}

// Helper functions

// HashPassword hashes the given password using bcrypt.
func (s *service) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePasswords compares a hashed password with a plain text password.
func (s *service) ComparePasswords(storedPassword string, inputPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(inputPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil // Passwords do not match
		}
		return false, err // Other error
	}
	return true, nil // Passwords match
}

/**
* Create Default Members.
**/
func (s *service) CreateDefaultMembers(members []CreateDefaultMember) error {
	var hashedPwMembers []CreateDefaultMember

	// Update members passwords with hash
	for _, member := range members {
		hashedPw, err := s.HashPassword(member.Password)
		if err != nil {
			return err
		}
		member.Password = hashedPw
		hashedPwMembers = append(hashedPwMembers, member)
	}

	return s.Repo.CreateDefaultMembers(context.Background(), hashedPwMembers)
}

// UpdateAvatarURL updates the member's avatar URL
func (s *service) UpdateAvatarURL(ctx context.Context, memberID uuid.UUID, avatarURL string) (*models.Member, error) {
	return s.Repo.UpdateAvatarURL(ctx, memberID, avatarURL)
}

// UpdateAvatarURL updates the member's avatar URL
func (s *service) UpdateAvatarURLTx(ctx context.Context, tx *sqlx.Tx, memberID uuid.UUID, avatarURL string) (*models.Member, error) {
	return s.Repo.UpdateAvatarURLTx(ctx, tx, memberID, avatarURL)
}

// SetStripeCustomerID saves a Stripe customer ID for the member
func (s *service) SetStripeCustomerID(ctx context.Context, req *pb.SetStripeCustomerIDRequest) (*pb.SetStripeCustomerIDResponse, error) {
	memberID, err := uuid.Parse(req.MemberId)
	if err != nil {
		return nil, fmt.Errorf("invalid member UUID: %w", err)
	}

	if err := s.Repo.SetStripeCustomerID(ctx, memberID, req.StripeCustomerId); err != nil {
		return nil, err
	}

	return &pb.SetStripeCustomerIDResponse{Success: true}, nil
}

// GetStripeCustomerID retrieves the Stripe customer ID for a member
func (s *service) GetStripeCustomerID(ctx context.Context, req *pb.GetStripeCustomerIDRequest) (*pb.GetStripeCustomerIDResponse, error) {
	memberID, err := uuid.Parse(req.MemberId)
	if err != nil {
		return nil, fmt.Errorf("invalid member UUID: %w", err)
	}

	customerID, err := s.Repo.GetStripeCustomerID(ctx, memberID)
	if err != nil {
		return nil, err
	}

	return &pb.GetStripeCustomerIDResponse{StripeCustomerId: customerID}, nil
}

// UpdateSubscriptionStatus updates the member's subscription product and status
func (s *service) UpdateSubscriptionStatus(ctx context.Context, req *pb.UpdateSubscriptionStatusRequest) (*pb.UpdateSubscriptionStatusResponse, error) {
	memberID, err := uuid.Parse(req.MemberId)
	if err != nil {
		return nil, fmt.Errorf("invalid member UUID: %w", err)
	}

	if err := s.Repo.UpdateSubscriptionStatus(ctx, memberID, req.ProductId, req.Status); err != nil {
		return nil, err
	}

	return &pb.UpdateSubscriptionStatusResponse{Success: true}, nil
}

func (s *service) CheckEmailExists(ctx context.Context, req *pb.CheckEmailRequest) (*pb.CheckEmailResponse, error) {
	_, err := s.Repo.GetMemberByEmail(ctx, req.Email)
	if err != nil {
		return &pb.CheckEmailResponse{Exists: false}, nil
	}
	return &pb.CheckEmailResponse{Exists: true}, nil
}
