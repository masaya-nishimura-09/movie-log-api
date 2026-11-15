package user

import (
	"context"
	"time"
)

type Repository interface {
	GetByID(ctx context.Context, userID ID) (*User, error)
	GetByEmail(ctx context.Context, email Email) (*User, error)
	Create(ctx context.Context, u *User) error
	ReplaceUnverified(ctx context.Context, u *User) error
	UpdateUsername(ctx context.Context, u *User) error
	UpdateEmail(ctx context.Context, u *User) error
	UpdatePassword(ctx context.Context, u *User) error
	UpdateEmailVerifiedAt(ctx context.Context, u *User) error
	Delete(ctx context.Context, userID ID) error
	DeleteUnverifiedCreatedBefore(ctx context.Context, before time.Time) error
}
