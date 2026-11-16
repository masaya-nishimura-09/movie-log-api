package user

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

type PasswordResetTokenRepository interface {
	Create(ctx context.Context, userID ID, language mail.Language) (*PasswordResetToken, error)
	FindValidByValue(ctx context.Context, value PasswordResetTokenValue) (*PasswordResetToken, error)
	FindLatestByUserID(ctx context.Context, userID ID) (*PasswordResetToken, error)
	DeleteAllForUser(ctx context.Context, userID ID) error
	DeleteExpired(ctx context.Context) error
}
