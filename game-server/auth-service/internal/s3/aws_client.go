package s3

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// AWSClient implements Client interface using AWS SDK v2
type AWSClient struct {
	client *s3.Client
	bucket string
}

// NewAWSClient creates a new AWS S3 client
func NewAWSClient(cfg aws.Config, bucket string) *AWSClient {
	return &AWSClient{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}
}

// GeneratePresignedUploadURL creates a presigned URL for uploading
func (c *AWSClient) GeneratePresignedUploadURL(ctx context.Context, key string, contentType string, expires time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(c.client)

	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	request, err := presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expires
	})

	if err != nil {
		return "", fmt.Errorf("generating presigned URL: %w", err)
	}

	return request.URL, nil
}

// HeadObject retrieves object metadata without downloading the object
func (c *AWSClient) HeadObject(ctx context.Context, key string) (*ObjectMetadata, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}

	result, err := c.client.HeadObject(ctx, input)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, fmt.Errorf("object not found: %w", err)
		}
		return nil, fmt.Errorf("retrieving object metadata: %w", err)
	}

	metadata := &ObjectMetadata{
		Key:         key,
		Size:        aws.ToInt64(result.ContentLength),
		ContentType: aws.ToString(result.ContentType),
	}

	if result.LastModified != nil {
		metadata.LastModified = *result.LastModified
	}

	return metadata, nil
}
