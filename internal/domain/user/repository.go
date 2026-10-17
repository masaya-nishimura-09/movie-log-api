package user

import (
	"context"
)

type Repository interface {
	GetByID(ctx context.Context, userID ID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	Create(ctx context.Context, u *User) error
	UpdateUsername(ctx context.Context, u *User) error
	UpdateEmail(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, u *User) error
	Delete(ctx context.Context, userID ID) error
}
