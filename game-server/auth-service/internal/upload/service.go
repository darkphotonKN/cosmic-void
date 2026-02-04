package upload

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// S3Client interface defines what the upload service needs from S3
// Following ISP - service owns its interface
type S3Client interface {
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expires time.Duration) (string, error)
	HeadObject(ctx context.Context, key string) (*ObjectMetadata, error)
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
	UpdateAvatarURL(ctx context.Context, memberID uuid.UUID, avatarURL string) error
}

// Service interface defines business logic operations
type Service interface {
	RequestAvatarUpload(ctx context.Context, memberID uuid.UUID, filename string) (*UploadRequest, error)
	ConfirmAvatarUpload(ctx context.Context, uploadID uuid.UUID) error
}

// service implements Service interface
type service struct {
	repo          Repository
	s3Client      S3Client
	memberService MemberService
	bucketName    string
	cdnURL        string // Optional CDN URL for serving images
	logger        *slog.Logger
}

// NewService creates a new upload service
func NewService(
	repo Repository,
	s3Client S3Client,
	memberService MemberService,
	bucketName string,
	cdnURL string,
	logger *slog.Logger,
) Service {
	return &service{
		repo:          repo,
		s3Client:      s3Client,
		memberService: memberService,
		bucketName:    bucketName,
		cdnURL:        cdnURL,
		logger:        logger,
	}
}

// RequestAvatarUpload creates a presigned URL for avatar upload
func (s *service) RequestAvatarUpload(ctx context.Context, memberID uuid.UUID, filename string) (*UploadRequest, error) {
	// Validate file extension
	ext := filepath.Ext(filename)
	contentType := s.getContentType(ext)
	if contentType == "" {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	// Generate unique S3 key
	timestamp := time.Now().Unix()
	randomID := uuid.New().String()[:8]
	s3Key := fmt.Sprintf("avatars/%s/%d_%s%s", memberID.String(), timestamp, randomID, ext)

	// Generate presigned URL
	expiresIn := 5 * time.Minute
	presignedURL, err := s.s3Client.GeneratePresignedUploadURL(ctx, s3Key, contentType, expiresIn)
	if err != nil {
		return nil, fmt.Errorf("generating presigned URL: %w", err)
	}

	// Create upload record
	uploadID := uuid.New()
	expiresAt := time.Now().Add(expiresIn)
	upload := &AvatarUpload{
		ID:                    uploadID,
		MemberID:              memberID,
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
		slog.String("member_id", memberID.String()),
		slog.String("s3_key", s3Key),
	)

	return &UploadRequest{
		UploadID:     uploadID,
		PresignedURL: presignedURL,
		S3Key:        s3Key,
		ExpiresAt:    expiresAt,
		MaxFileSize:  MaxAvatarSize,
		AllowedContentTypes: []string{
			ContentTypeJPEG,
			ContentTypePNG,
			ContentTypeWEBP,
		},
	}, nil
}

// ConfirmAvatarUpload confirms successful upload and updates member avatar URL
func (s *service) ConfirmAvatarUpload(ctx context.Context, uploadID uuid.UUID) error {
	// Get upload record
	upload, err := s.repo.GetUploadByID(ctx, uploadID)
	if err != nil {
		return fmt.Errorf("retrieving upload: %w", err)
	}

	// Check if already processed
	if upload.UploadStatus != StatusPending {
		return fmt.Errorf("upload already processed with status: %s", upload.UploadStatus)
	}

	// Verify object exists in S3
	metadata, err := s.s3Client.HeadObject(ctx, upload.S3Key)
	if err != nil {
		// Mark as failed
		_ = s.repo.UpdateUploadStatus(ctx, uploadID, StatusFailed)
		return fmt.Errorf("verifying S3 object: %w", err)
	}

	// Validate file size
	if metadata.Size > MaxAvatarSize {
		_ = s.repo.UpdateUploadStatus(ctx, uploadID, StatusFailed)
		return fmt.Errorf("file size %d exceeds maximum %d", metadata.Size, MaxAvatarSize)
	}

	// Update upload record with file metadata
	upload.FileSize = &metadata.Size
	if err := s.repo.UpdateUploadStatus(ctx, uploadID, StatusUploaded); err != nil {
		return fmt.Errorf("updating upload status: %w", err)
	}

	// Generate avatar URL (use CDN if configured, otherwise S3 URL)
	var avatarURL string
	if s.cdnURL != "" {
		avatarURL = fmt.Sprintf("%s/%s", s.cdnURL, upload.S3Key)
	} else {
		avatarURL = fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucketName, upload.S3Key)
	}

	// Update member avatar URL
	if err := s.memberService.UpdateAvatarURL(ctx, upload.MemberID, avatarURL); err != nil {
		return fmt.Errorf("updating member avatar URL: %w", err)
	}

	// Mark as synced
	if err := s.repo.UpdateUploadStatus(ctx, uploadID, StatusSynced); err != nil {
		return fmt.Errorf("marking upload as synced: %w", err)
	}

	s.logger.InfoContext(ctx, "avatar upload confirmed",
		slog.String("upload_id", uploadID.String()),
		slog.String("member_id", upload.MemberID.String()),
		slog.String("avatar_url", avatarURL),
		slog.Int64("file_size", metadata.Size),
	)

	return nil
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

