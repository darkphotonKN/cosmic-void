package upload

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/models"
	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/s3"
	pb "github.com/darkphotonKN/cosmic-void-server/common/api/proto/auth"
	pbevents "github.com/darkphotonKN/cosmic-void-server/common/api/proto/events"
	commonbroker "github.com/darkphotonKN/cosmic-void-server/common/broker"
	commonconstants "github.com/darkphotonKN/cosmic-void-server/common/constants"
	commonutils "github.com/darkphotonKN/cosmic-void-server/common/utils"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// S3Client interface defines what the upload service needs from S3
// Following ISP - service owns its interface
type S3Client interface {
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expires time.Duration) (string, error)
	HeadObject(ctx context.Context, key string) (*s3.ObjectMetadata, error)
}

// ObjectMetadata contains metadata about an S3 object
type ObjectMetadata struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

// MemberService interface defines what the upload service needs from member service
type MemberService interface {
	UpdateAvatarURLTx(ctx context.Context, tx *sqlx.Tx, memberID uuid.UUID, avatarURL string) (*models.Member, error)
}

// Repository interface defines data access operations
// Following ISP - service defines what it needs from repository
type Repository interface {
	CreateUpload(ctx context.Context, upload *AvatarUpload) error
	GetUploadByID(ctx context.Context, id uuid.UUID) (*AvatarUpload, error)
	UpdateUploadStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateUploadStatusTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status string) error
	GetPendingUploadsByMember(ctx context.Context, memberID uuid.UUID) ([]*AvatarUpload, error)
}

// service implements Service interface
type service struct {
	repo          Repository
	s3Client      S3Client
	memberService MemberService
	bucketName    string
	cdnURL        string // Optional CDN URL for serving images
	logger        *slog.Logger
	publishCh     commonbroker.Publisher
	db            *sqlx.DB
}

// NewService creates a new upload service
func NewService(
	repo Repository,
	s3Client S3Client,
	memberService MemberService,
	bucketName string,
	cdnURL string,
	logger *slog.Logger,
	pubishCh commonbroker.Publisher,
	db *sqlx.DB,
) *service {
	return &service{
		repo:          repo,
		s3Client:      s3Client,
		memberService: memberService,
		bucketName:    bucketName,
		cdnURL:        cdnURL,
		logger:        logger,
		publishCh:     pubishCh,
		db:            db,
	}
}

// RequestAvatarUpload creates a presigned URL for avatar upload
func (s *service) RequestAvatarUpload(ctx context.Context, req *pb.RequestAvatarUploadRequest) (*pb.RequestAvatarUploadResponse, error) {
	// Validate file extension
	ext := filepath.Ext(req.Filename)
	contentType := s.getContentType(ext)
	if contentType == "" {
		slog.Error("Attempted upload with unsupported file type.")
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	// Generate unique S3 key
	timestamp := time.Now().Unix()
	randomID := uuid.New().String()[:8]
	s3Key := fmt.Sprintf("avatars/%s/%d_%s%s", req.MemberId, timestamp, randomID, ext)

	// Generate presigned URL
	expiresIn := 5 * time.Minute
	presignedURL, err := s.s3Client.GeneratePresignedUploadURL(ctx, s3Key, contentType, expiresIn)

	if err != nil {
		slog.Error("Error when attempting to generate presigned upload url", "err", err)
		return nil, fmt.Errorf("generating presigned URL: %w", err)
	}

	memberId, err := uuid.Parse(req.MemberId)

	if err != nil {
		slog.Error("Error when parsing memberId from proto request into uuid.", "err", err)
		return nil, err
	}

	// Create upload record
	uploadID := uuid.New()
	expiresAt := time.Now().Add(expiresIn)
	upload := &AvatarUpload{
		ID:                    uploadID,
		MemberID:              memberId,
		S3Key:                 s3Key,
		UploadStatus:          StatusPending,
		ContentType:           &contentType,
		PresignedURLExpiresAt: &expiresAt,
	}

	if err := s.repo.CreateUpload(ctx, upload); err != nil {
		return nil, fmt.Errorf("creating upload record: %w", err)
	}

	s.logger.InfoContext(ctx, "avatar upload requested",
		slog.String("upload_id", uploadID.String()),
		slog.String("member_id", req.MemberId),
		slog.String("s3_key", s3Key),
	)

	res := &pb.RequestAvatarUploadResponse{
		UploadId:     uploadID.String(),
		PresignedUrl: presignedURL,
		S3Key:        s3Key,
		ExpiresAt:    timestamppb.New(expiresAt),
		MaxFileSize:  MaxAvatarSize,
		AllowedContentTypes: []string{
			ContentTypeJPEG,
			ContentTypePNG,
			ContentTypeWEBP,
		},
	}

	return res, nil
}

// ConfirmAvatarUpload confirms successful upload and updates member avatar URL
func (s *service) ConfirmAvatarUpload(ctx context.Context, req *pb.ConfirmAvatarUploadRequest) (*pb.ConfirmAvatarUploadResponse, error) {
	uploadId, err := uuid.Parse(req.UploadId)

	if err != nil {
		slog.Error("Error when parsing memberId from proto request into uuid.", "err", err)
		return nil, err
	}

	// Get upload record
	upload, err := s.repo.GetUploadByID(ctx, uploadId)
	if err != nil {
		return nil, fmt.Errorf("retrieving upload: %w", err)
	}

	// Check if already processed
	if upload.UploadStatus != StatusPending {
		return nil, fmt.Errorf("upload already processed with status: %s", upload.UploadStatus)
	}

	// Verify object exists in S3
	metadata, err := s.s3Client.HeadObject(ctx, upload.S3Key)
	if err != nil {
		// Mark as failed
		_ = s.repo.UpdateUploadStatus(ctx, uploadId, StatusFailed)
		return nil, fmt.Errorf("verifying S3 object: %w", err)
	}

	// Validate file size
	if metadata.Size > MaxAvatarSize {
		_ = s.repo.UpdateUploadStatus(ctx, uploadId, StatusFailed)
		return nil, fmt.Errorf("file size %d exceeds maximum %d", metadata.Size, MaxAvatarSize)
	}

	// Update upload record with file metadata
	upload.FileSize = &metadata.Size
	if err := s.repo.UpdateUploadStatus(ctx, uploadId, StatusUploaded); err != nil {
		return nil, fmt.Errorf("updating upload status: %w", err)
	}

	// Generate avatar URL (use CDN if configured, otherwise S3 URL)
	var avatarURL string
	if s.cdnURL != "" {
		avatarURL = fmt.Sprintf("%s/%s", s.cdnURL, upload.S3Key)
	} else {
		avatarURL = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, upload.S3Key)
	}

	var updatedMember *models.Member

	err = commonutils.ExecTx(ctx, s.db, func(tx *sqlx.Tx) error {
		// Update member avatar URL
		res, err := s.memberService.UpdateAvatarURLTx(ctx, tx, upload.MemberID, avatarURL)
		if err != nil {
			return fmt.Errorf("updating member avatar URL: %w", err)
		}

		s.logger.Debug("after UpdateAvatarURL call", "updatedMember", updatedMember)

		// Mark as synced
		if err := s.repo.UpdateUploadStatusTx(ctx, tx, uploadId, StatusSynced); err != nil {
			return fmt.Errorf("marking upload as synced: %w", err)
		}

		// everything succeeded, update outerscope return values
		updatedMember = res

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("confirming avatar upload: %w", err)
	}

	s.logger.InfoContext(ctx, "avatar upload confirmed",
		slog.String("upload_id", uploadId.String()),
		slog.String("member_id", upload.MemberID.String()),
		slog.String("avatar_url", avatarURL),
		slog.Int64("file_size", metadata.Size),
	)

	// confirmed, fire off amqp event for profile sync (dernormalized ranking leaderboard tables)

	protoData, err := proto.Marshal(&pbevents.MemberProfileUpdatedEvent{
		MemberId:  upload.MemberID.String(),
		Username:  updatedMember.Name,
		AvatarUrl: avatarURL,
	})

	s.publishCh.PublishWithContext(ctx, commonconstants.AuthEventsExchange, commonconstants.MemberProfileUpdated,
		commonbroker.Message{
			ContentType:  "application/protobuf",
			Body:         protoData,
			DeliveryMode: amqp.Persistent,
		},
	)

	return &pb.ConfirmAvatarUploadResponse{
		Success:   true,
		Message:   "Avatar was confirmed to be uploaded.",
		AvatarUrl: avatarURL,
	}, nil
}

// getContentType returns content type based on file extension
func (s *service) getContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return ContentTypeJPEG
	case ".png":
		return ContentTypePNG
	case ".webp":
		return ContentTypeWEBP
	default:
		return ""
	}
}
