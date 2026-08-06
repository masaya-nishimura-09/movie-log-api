package user

import (
	"context"
)

type UserRepository interface {
	GetByID(ctx context.Context, userID ID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, userID ID) error
}
