package config

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3Client() (*s3.Client, error) {
	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	if cfg.Region == "" {
		return nil, fmt.Errorf("environment variable AWS_REGION is required")
	}

	if _, err := cfg.Credentials.Retrieve(ctx); err != nil {
		return nil, fmt.Errorf("failed to retrieve AWS credentials: %w", err)
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

func S3PublicBaseURL() (string, error) {
	baseURL := os.Getenv("S3_PUBLIC_BASE_URL")
	if baseURL == "" {
		return "", fmt.Errorf("environment variable S3_PUBLIC_BASE_URL is required")
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid S3_PUBLIC_BASE_URL: %w", err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("environment variable S3_PUBLIC_BASE_URL must be an http or https url")
	}

	return baseURL, nil
}
