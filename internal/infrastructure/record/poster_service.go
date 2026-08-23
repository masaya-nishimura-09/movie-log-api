package record

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type posterService struct {
	client  *s3.Client
	bucket  string
	baseURL string
}

func NewPosterService(client *s3.Client, bucket string, baseURL string) record.PosterService {
	return &posterService{client: client, bucket: bucket, baseURL: baseURL}
}

func extension(contentType record.PosterContentType) string {
	switch contentType {
	case record.PosterContentTypeJPEG:
		return ".jpg"
	case record.PosterContentTypePNG:
		return ".png"
	case record.PosterContentTypeWebP:
		return ".webp"
	default:
		return ""
	}
}

func (ps *posterService) Upload(
	ctx context.Context,
	userID user.ID,
	poster record.Poster,
) (record.PosterURL, error) {
	key := fmt.Sprintf("%d/%s%s", uint(userID), uuid.NewString(), extension(poster.ContentType))

	_, err := ps.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(ps.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(poster.Data),
		ContentType: aws.String(string(poster.ContentType)),
	})
	if err != nil {
		return "", fmt.Errorf("upload poster: %w", err)
	}

	return record.PosterURL(ps.baseURL + "/" + key), nil
}

func (ps *posterService) Delete(
	ctx context.Context,
	userID user.ID,
	url record.PosterURL,
) error {
	prefix := ps.baseURL + "/"
	if !strings.HasPrefix(string(url), prefix) {
		return nil
	}

	key := strings.TrimPrefix(string(url), prefix)
	if !strings.HasPrefix(key, fmt.Sprintf("%d/", uint(userID))) {
		return nil
	}

	_, err := ps.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(ps.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete poster: %w", err)
	}
	return nil
}
