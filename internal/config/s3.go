package config

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3Client() (*s3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	endpoint := os.Getenv("S3_ENDPOINT")
	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint == "" {
			return
		}
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	}), nil
}

func S3Bucket() (string, error) {
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return "", fmt.Errorf("environment variable S3_BUCKET is required")
	}
	return bucket, nil
}
