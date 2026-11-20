package auth

import (
	"context"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, principal *Principal) (*RefreshToken, error)
	CreateWithExpiresAt(
		ctx context.Context,
		principal *Principal,
		expiresAt time.Time,
	) (*RefreshToken, error)
	FindValidByValue(
		ctx context.Context,
		value RefreshTokenValue,
	) (*RefreshToken, error)
	FindRefreshableByValue(
		ctx context.Context,
		value RefreshTokenValue,
	) (*RefreshToken, error)
	Revoke(ctx context.Context, id RefreshTokenID) error
	Rotate(ctx context.Context, id RefreshTokenID) error
	RevokeAllForUser(ctx context.Context, userID user.ID) error
	DeleteExpired(ctx context.Context) error
}
