package config

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/darkphotonKN/cosmic-void-server/auth-service/internal/s3"
	commonhelpers "github.com/darkphotonKN/cosmic-void-server/common/utils"
)

// S3Config holds S3 configuration
type S3Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	CDNUrl          string
}

// LoadS3Config loads S3 configuration from environment variables
func LoadS3Config() S3Config {
	return S3Config{
		Region:          commonhelpers.GetEnvString("AWS_REGION", "us-east-1"),
		AccessKeyID:     commonhelpers.GetEnvString("AWS_ACCESS_KEY_ID", ""),
		SecretAccessKey: commonhelpers.GetEnvString("AWS_SECRET_ACCESS_KEY", ""),
		BucketName:      commonhelpers.GetEnvString("S3_BUCKET_NAME", ""),
		CDNUrl:          commonhelpers.GetEnvString("CDN_URL", ""),
	}
}

// InitS3Client initializes the S3 client with AWS configuration
func InitS3Client(ctx context.Context, cfg S3Config) (s3.Client, error) {
	// Validate required configuration
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("AWS credentials not configured")
	}
	if cfg.BucketName == "" {
		return nil, fmt.Errorf("S3 bucket name not configured")
	}

	// Create AWS config with static credentials
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	// Create and return S3 client
	s3Client := s3.NewAWSClient(awsCfg, cfg.BucketName)

	slog.Info("S3 client initialized",
		slog.String("region", cfg.Region),
		slog.String("bucket", cfg.BucketName),
		slog.Bool("cdn_configured", cfg.CDNUrl != ""),
	)

	return s3Client, nil
}

// GetAWSConfig returns AWS configuration for other AWS services if needed
func GetAWSConfig(ctx context.Context, cfg S3Config) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			),
		),
	)
}

