package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

type accessTokenService struct {
	secret []byte
	ttl    time.Duration
}

func NewAccessTokenService(secret []byte, ttl time.Duration) auth.AccessTokenService {
	return &accessTokenService{secret: secret, ttl: ttl}
}

type claims struct {
	UserID user.ID
	Role   user.Role
	jwt.RegisteredClaims
}

func (r *accessTokenService) Generate(
	ctx context.Context,
	principal *auth.Principal,
) (*auth.AccessToken, error) {
	c := claims{
		UserID: principal.UserID,
		Role:   principal.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(r.ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenStr, err := t.SignedString(r.secret)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	accessToken := auth.AccessToken{Value: auth.AccessTokenValue(tokenStr)}
	return &accessToken, nil
}

func (r *accessTokenService) Validate(
	ctx context.Context,
	accessToken *auth.AccessToken,
) (
	*auth.Principal,
	error,
) {
	parsedToken, err := jwt.ParseWithClaims(
		string(accessToken.Value),
		&claims{},
		func(t *jwt.Token) (any, error) {
			return r.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}
	if !parsedToken.Valid {
		return nil, exception.ErrInvalid
	}
	principal := auth.Principal{
		UserID: parsedToken.Claims.(*claims).UserID,
		Role:   parsedToken.Claims.(*claims).Role,
	}
	return &principal, nil
}
