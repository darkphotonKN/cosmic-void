package upload

import (
	"time"

	"github.com/google/uuid"
)

// AvatarUpload represents an avatar upload record
type AvatarUpload struct {
	ID                     uuid.UUID  `db:"id"`
	MemberID               uuid.UUID  `db:"member_id"`
	S3Key                  string     `db:"s3_key"`
	UploadStatus           string     `db:"upload_status"`
	FileSize               *int64     `db:"file_size"`
	ContentType            *string    `db:"content_type"`
	PresignedURLExpiresAt  *time.Time `db:"presigned_url_expires_at"`
	CreatedAt              time.Time  `db:"created_at"`
	UpdatedAt              time.Time  `db:"updated_at"`
}

// UploadRequest represents a request for an avatar upload
type UploadRequest struct {
	UploadID              uuid.UUID `json:"upload_id"`
	PresignedURL          string    `json:"presigned_url"`
	S3Key                 string    `json:"s3_key"`
	ExpiresAt             time.Time `json:"expires_at"`
	MaxFileSize           int64     `json:"max_file_size"`
	AllowedContentTypes   []string  `json:"allowed_content_types"`
}

// UploadConfirmation represents confirmation of a successful upload
type UploadConfirmation struct {
	UploadID uuid.UUID `json:"upload_id"`
}

// Upload status constants
const (
	StatusPending  = "pending"
	StatusUploaded = "uploaded"
	StatusSynced   = "synced"
	StatusFailed   = "failed"
)

// Content type constants
const (
	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"
	ContentTypeWEBP = "image/webp"
)

// Size limits
const (
	MaxAvatarSize = 5 * 1024 * 1024 // 5MB
)