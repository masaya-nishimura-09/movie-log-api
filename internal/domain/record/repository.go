package record

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type RecordRepository interface {
	GetByID(ctx context.Context, recordID ID) (*Record, error)
	ListByUserID(ctx context.Context, userID user.ID) ([]*Record, error)
	Create(ctx context.Context, r *Record) error
	Update(ctx context.Context, r *Record) error
	Delete(ctx context.Context, recordID ID) error
}
