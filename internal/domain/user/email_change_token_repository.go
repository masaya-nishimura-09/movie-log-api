package user

import (
	"context"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
)

type EmailChangeTokenRepository interface {
	Create(ctx context.Context, userID ID, newEmail Email, language mail.Language) (*EmailChangeToken, error)
	FindValidByValue(ctx context.Context, value EmailChangeTokenValue) (*EmailChangeToken, error)
	DeleteAllForUser(ctx context.Context, userID ID) error
	DeleteExpired(ctx context.Context) error
}
