package s3

import (
	"context"
	"time"
)

// ObjectMetadata contains metadata about an S3 object
type ObjectMetadata struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
}

// Client interface defines S3 operations needed by consumers
// Following ISP - only expose what consumers need
type Client interface {
	// GeneratePresignedUploadURL creates a presigned URL for uploading
	GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expires time.Duration) (string, error)

	// HeadObject retrieves object metadata without downloading the object
	HeadObject(ctx context.Context, key string) (*ObjectMetadata, error)
}

