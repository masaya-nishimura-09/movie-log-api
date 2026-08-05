package repository

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/model"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/repository"
)

type tokenRepository struct {
	secret []byte
}

type claims struct {
	UserID model.UserID
	Role   model.Role
	jwt.RegisteredClaims
}

func NewTokenRepo(secret []byte) repository.TokenRepository {
	return &tokenRepository{secret}
}

func (tr *tokenRepository) Generate(principal *model.Principal) (model.Token, error) {
	c := claims{
		UserID: principal.UserID,
		Role:   principal.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenStr, err := t.SignedString(tr.secret)
	if err != nil {
		return model.Token(""), fmt.Errorf("generate token: %w", err)
	}
	return model.Token(tokenStr), nil
}

func (tr *tokenRepository) Validate(token model.Token) (*model.Principal, error) {
	parsedToken, err := jwt.ParseWithClaims(string(token), &claims{}, func(t *jwt.Token) (any, error) {
		return tr.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}
	if !parsedToken.Valid {
		return nil, model.ErrInvalidToken
	}
	principal := model.Principal{UserID: parsedToken.Claims.(*claims).UserID, Role: parsedToken.Claims.(*claims).Role}
	return &principal, nil
}
