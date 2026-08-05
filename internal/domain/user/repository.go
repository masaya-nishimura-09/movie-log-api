package user

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/model"
)

type UserRepository interface {
	GetByID(ctx context.Context, userID model.UserID) (*model.User, error)
	GetByEmail(ctx context.Context, email model.Email) (*model.User, error)
	Create(ctx context.Context, u *model.User) error
	Update(ctx context.Context, u *model.User) error
	UpdatePassword(ctx context.Context, u *model.User) error
	Delete(ctx context.Context, userID model.UserID) error
}
