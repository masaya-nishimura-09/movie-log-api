package media

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type Service interface {
	Upload(ctx context.Context, userID user.ID, media Media) (URL, error)
	Delete(ctx context.Context, userID user.ID, url URL) error
	DeleteAllForUser(ctx context.Context, userID user.ID) error
}
