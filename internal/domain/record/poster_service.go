package record

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type PosterService interface {
	Upload(ctx context.Context, userID user.ID, poster Poster) (PosterURL, error)
	Delete(ctx context.Context, url PosterURL) error
}
