package user

import (
	"context"
)

type EmailVerificationTokenRepository interface {
	Create(ctx context.Context, userID ID) (*EmailVerificationToken, error)
	FindValidByValue(ctx context.Context, value EmailVerificationTokenValue) (*EmailVerificationToken, error)
	FindLatestByUserID(ctx context.Context, userID ID) (*EmailVerificationToken, error)
	DeleteAllForUser(ctx context.Context, userID ID) error
}
