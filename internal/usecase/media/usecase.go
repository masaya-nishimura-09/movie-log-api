package media

import (
	"context"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/media"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type Usecase interface {
	Upload(ctx context.Context, userID user.ID, m media.Media) (media.URL, error)
	Delete(ctx context.Context, userID user.ID, url media.URL) error
}

type MediaUsecase struct {
	mediaService media.Service
}

func NewMediaUsecase(mediaService media.Service) *MediaUsecase {
	return &MediaUsecase{mediaService: mediaService}
}

func (mu *MediaUsecase) Upload(
	ctx context.Context,
	userID user.ID,
	m media.Media,
) (media.URL, error) {
	url, err := mu.mediaService.Upload(ctx, userID, m)
	if err != nil {
		return "", fmt.Errorf("upload media: %w", err)
	}
	return url, nil
}

func (mu *MediaUsecase) Delete(
	ctx context.Context,
	userID user.ID,
	url media.URL,
) error {
	if err := mu.mediaService.Delete(ctx, userID, url); err != nil {
		return fmt.Errorf("delete media: %w", err)
	}
	return nil
}
