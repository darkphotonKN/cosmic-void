package member

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/auth"
	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/models"
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	"github.com/darkphotonKN/cosmic-void-server/common/utils/cache"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type service struct {
	Repo      *Repository
	publishCh *amqp.Channel
	cache     cache.Cache
}

func NewService(repo *Repository, ch *amqp.Channel, cacheService cache.Cache) *service {
	return &service{
		Repo:      repo,
		publishCh: ch,
		cache:     cacheService,
	}
}

func memberToProto(m *models.Member) *pb.Member {
	if m == nil {
		return nil
	}

	return &pb.Member{
		Id:            m.ID.String(),
		Name:          m.Name,
		Email:         m.Email,
		Status:        int32(stringToInt(m.Status)),
		AverageRating: float32(m.AverageRating),
		CreatedAt:     timestamppb.New(m.CreatedAt),
		UpdatedAt:     timestamppb.New(m.UpdatedAt),
	}
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

	member, err := s.Repo.GetById(id)
	if err != nil {
		return nil, err
	}

	return memberToProto(member), nil
}

func (s *service) CreateMember(ctx context.Context, req *pb.CreateMemberRequest) (*pb.Member, error) {
	// Hash the password
	hashedPw, err := s.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create the member
	memberId, err := s.Repo.Create(req.Name, req.Email, hashedPw)
	if err != nil {
		return nil, err
	}

	// publish to message broker
	payload := commonconstants.MemberSignedUpEventPayload{
		UserID:     memberId.String(),
		Name:       req.Name,
		Email:      req.Email,
		SignedUpAt: "", // TODO: update this to legit date
	}

	marshalledPayload, err := json.Marshal(payload)

	if err != nil {
		return nil, err
	}

	err = s.publishCh.PublishWithContext(
		ctx,
		commonconstants.MemberSignedUpEvent,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        marshalledPayload,
			// persist message
			DeliveryMode: amqp.Persistent,
		})

	// Get the created member
	member, err := s.Repo.GetById(memberId)
	if err != nil {
		return nil, err
	}

	return memberToProto(member), nil
}

func (s *service) LoginMember(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	member, err := s.Repo.GetMemberByEmail(req.Email)
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

	fmt.Println("generated tokens:", accessToken)

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
	err = s.Repo.UpdateMemberInfo(id, req.Name, req.Status)
	if err != nil {
		return nil, err
	}

	// Get the updated member
	member, err := s.Repo.GetById(id)
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
	member, err := s.Repo.GetByIdWithPassword(id)
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

	err = s.Repo.UpdatePassword(params)
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

	return s.Repo.CreateDefaultMembers(hashedPwMembers)
}
