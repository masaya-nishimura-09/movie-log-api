package auth

import "context"

type AccessTokenService interface {
	Generate(ctx context.Context, principal *Principal) (*AccessToken, error)
	Validate(ctx context.Context, accessToken *AccessToken) (*Principal, error)
}
