package auth

import "context"

type RefreshTokenRepository interface {
	Create(ctx context.Context, principal *Principal) (*RefreshToken, error)
	FindValidByValue(
		ctx context.Context,
		value RefreshTokenValue,
	) (*RefreshToken, error)
	Revoke(ctx context.Context, id RefreshTokenID) error
}
