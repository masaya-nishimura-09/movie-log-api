package testutil

import (
	"log"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/masaya-nishimura-09/movie-log-api/internal/config"
)

func NewTestS3Client() *s3.Client {
	loadTestEnv()

	client, err := config.NewS3Client()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return client
}

func S3Bucket() string {
	loadTestEnv()

	bucket, err := config.S3Bucket()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return bucket
}

func S3BaseURL() string {
	loadTestEnv()

	baseURL, err := config.S3PublicBaseURL()
	if err != nil {
		log.Fatalf("%v", err)
	}

	return baseURL
}
