package media

import (
	"context"
	"errors"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type fakeService struct {
	url            media.URL
	err            error
	uploadedUserID user.ID
	uploadedMedia  media.Media
	deletedUserID  user.ID
	deletedURL     media.URL
}

func (s *fakeService) Upload(
	ctx context.Context,
	userID user.ID,
	m media.Media,
) (media.URL, error) {
	s.uploadedUserID = userID
	s.uploadedMedia = m
	return s.url, s.err
}

func (s *fakeService) Delete(
	ctx context.Context,
	userID user.ID,
	url media.URL,
) error {
	s.deletedUserID = userID
	s.deletedURL = url
	return s.err
}

func (s *fakeService) DeleteAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	return s.err
}

func TestUpload(t *testing.T) {
	t.Run(
		"passes values to the service and returns the URL",
		func(t *testing.T) {
			url := media.URL("https://example.com/1/image.jpg")
			service := &fakeService{url: url}
			mu := NewMediaUsecase(service)

			ctx := context.Background()
			userID := user.ID(1)
			m := media.Media{
				Data:        media.Data([]byte("\xFF\xD8\xFF")),
				ContentType: media.ContentTypeJPEG,
			}

			got, err := mu.Upload(ctx, userID, m)
			if err != nil {
				t.Fatalf(
					"Upload(ctx, %v, media) (media.URL, error) = %v, %v",
					userID, got, err,
				)
			}
			if got != url {
				t.Errorf(
					"Upload(ctx, %v, media) URL = %v, want %v",
					userID, got, url,
				)
			}

			if service.uploadedUserID != userID {
				t.Errorf(
					"Upload(ctx, %v, media) service user id = %v, want %v",
					userID, service.uploadedUserID, userID,
				)
			}
			if service.uploadedMedia.ContentType != m.ContentType {
				t.Errorf(
					"Upload(ctx, %v, media) service content type = %v, want %v",
					userID, service.uploadedMedia.ContentType, m.ContentType,
				)
			}
		},
	)

	t.Run(
		"returns an error when the service fails",
		func(t *testing.T) {
			serviceErr := errors.New("upload failed")
			service := &fakeService{err: serviceErr}
			mu := NewMediaUsecase(service)

			ctx := context.Background()
			userID := user.ID(1)
			m := media.Media{
				Data:        media.Data([]byte("\xFF\xD8\xFF")),
				ContentType: media.ContentTypeJPEG,
			}

			got, err := mu.Upload(ctx, userID, m)
			if !errors.Is(err, serviceErr) {
				t.Errorf(
					"Upload(ctx, %v, media) error = %v, want %v",
					userID, err, serviceErr,
				)
			}
			if got != "" {
				t.Errorf(
					"Upload(ctx, %v, media) URL = %v, want empty",
					userID, got,
				)
			}
		},
	)
}

func TestDelete(t *testing.T) {
	t.Run(
		"passes values to the service",
		func(t *testing.T) {
			service := &fakeService{}
			mu := NewMediaUsecase(service)

			ctx := context.Background()
			userID := user.ID(1)
			url := media.URL("https://example.com/1/image.jpg")

			err := mu.Delete(ctx, userID, url)
			if err != nil {
				t.Fatalf(
					"Delete(ctx, %v, %v) error = %v",
					userID, url, err,
				)
			}

			if service.deletedUserID != userID {
				t.Errorf(
					"Delete(ctx, %v, %v) service user id = %v, want %v",
					userID, url, service.deletedUserID, userID,
				)
			}
			if service.deletedURL != url {
				t.Errorf(
					"Delete(ctx, %v, %v) service url = %v, want %v",
					userID, url, service.deletedURL, url,
				)
			}
		},
	)

	t.Run(
		"returns an error when the service fails",
		func(t *testing.T) {
			serviceErr := errors.New("delete failed")
			service := &fakeService{err: serviceErr}
			mu := NewMediaUsecase(service)

			ctx := context.Background()
			userID := user.ID(1)
			url := media.URL("https://example.com/1/image.jpg")

			err := mu.Delete(ctx, userID, url)
			if !errors.Is(err, serviceErr) {
				t.Errorf(
					"Delete(ctx, %v, %v) error = %v, want %v",
					userID, url, err, serviceErr,
				)
			}
		},
	)
}
