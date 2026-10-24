package record

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type ArticleRepository interface {
	GetByID(ctx context.Context, articleID ArticleID) (*ArticleDetail, error)
	ListByUserID(ctx context.Context, userID user.ID, query ArticleQuery) (ArticleListResult, error)
	Create(ctx context.Context, a *Article) error
	Update(ctx context.Context, a *Article) error
	Delete(ctx context.Context, articleID ArticleID) error
}
