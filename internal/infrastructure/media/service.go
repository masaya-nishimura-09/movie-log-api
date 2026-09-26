package media

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type service struct {
	client  *s3.Client
	bucket  string
	baseURL string
}

func NewService(client *s3.Client, bucket string, baseURL string) media.Service {
	return &service{client: client, bucket: bucket, baseURL: baseURL}
}

func extension(contentType media.ContentType) string {
	switch contentType {
	case media.ContentTypeJPEG:
		return ".jpg"
	case media.ContentTypePNG:
		return ".png"
	case media.ContentTypeWebP:
		return ".webp"
	default:
		return ""
	}
}

func (s *service) Upload(
	ctx context.Context,
	userID user.ID,
	m media.Media,
) (media.URL, error) {
	key := fmt.Sprintf("%d/%s%s", uint(userID), uuid.NewString(), extension(m.ContentType))

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(m.Data),
		ContentType: aws.String(string(m.ContentType)),
	})
	if err != nil {
		return "", fmt.Errorf("upload media: %w", err)
	}

	return media.URL(s.baseURL + "/" + key), nil
}

func (s *service) Delete(
	ctx context.Context,
	userID user.ID,
	url media.URL,
) error {
	prefix := s.baseURL + "/"
	if !strings.HasPrefix(string(url), prefix) {
		return nil
	}

	key := strings.TrimPrefix(string(url), prefix)
	if !strings.HasPrefix(key, fmt.Sprintf("%d/", uint(userID))) {
		return nil
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete media: %w", err)
	}
	return nil
}

func (s *service) DeleteAllForUser(ctx context.Context, userID user.ID) error {
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(fmt.Sprintf("%d/", uint(userID))),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list media: %w", err)
		}

		for _, object := range page.Contents {
			_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: aws.String(s.bucket),
				Key:    object.Key,
			})
			if err != nil {
				return fmt.Errorf("delete media: %w", err)
			}
		}
	}
	return nil
}
