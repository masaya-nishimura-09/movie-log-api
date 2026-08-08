package auth

import "context"

type RefreshTokenRepository interface {
	Create(ctx context.Context, principal *Principal) (*RefreshToken, error)
}
