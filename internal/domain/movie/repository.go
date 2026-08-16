package movie

import (
	"context"
)

type MovieRepository interface {
	GetByID(ctx context.Context, movieID ID) (*Movie, error)
	GetByTMDBID(ctx context.Context, tmdbID MovieTMDBID) (*Movie, error)
	Create(ctx context.Context, m *Movie) error
	Update(ctx context.Context, m *Movie) error
	Delete(ctx context.Context, movieID ID) error
}
