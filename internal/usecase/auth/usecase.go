package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo  user.UserRepository
	refreshTokenRepo auth.RefreshTokenRepository
	secret []byte
}

type claims struct {
	UserID user.ID
	Role   user.Role
	jwt.RegisteredClaims
}

func NewAuthUsecase(
	userRepo user.UserRepository, 
	refreshTokenRepo auth.RefreshTokenRepository,
	secret []byte,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo, 
		refreshTokenRepo: refreshTokenRepo,
		secret: secret,
	}
}

func generateAccessToken(
	principal *auth.Principal,
	secret []byte,
) (
	*auth.AccessToken, 
	error,
) {
	c := claims{
		UserID: principal.UserID,
		Role:   principal.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	tokenStr, err := t.SignedString(secret)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	accessToken := auth.AccessToken{Value: tokenStr}
	return &accessToken, nil
}

func (au *AuthUsecase) ValidateAccessToken(accessToken *auth.AccessToken) (
	*auth.Principal, 
	error,
) {
	parsedToken, err := jwt.ParseWithClaims(
		string(accessToken.Value), 
		&claims{}, 
		func(t *jwt.Token) (any, error) {
			return au.secret, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("validate token: %w", err)
	}
	if !parsedToken.Valid {
		return nil, exception.ErrInvalidToken
	}
	principal := auth.Principal{
		UserID: parsedToken.Claims.(*claims).UserID, 
		Role: parsedToken.Claims.(*claims).Role,
	}
	return &principal, nil
}

func (au *AuthUsecase) Login(
	ctx context.Context, 
	email user.Email, 
	password user.Password,
) (*auth.AccessToken, *auth.RefreshToken, error) {
	existingUser, err := au.userRepo.GetByEmail(ctx, email)
	if errors.Is(err, exception.ErrUserNotFound) {
		return nil, nil, exception.ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, fmt.Errorf("login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(existingUser.HashedPassword), 
		[]byte(password),
	); err != nil {
		return nil, nil, exception.ErrInvalidCredentials
	}

	principal := auth.Principal{
		UserID: user.ID(existingUser.ID),
		Role:   user.Role(existingUser.Role),
	}

	accessToken, err := generateAccessToken(&principal, au.secret)
	if err != nil {
		return nil, nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := au.refreshTokenRepo.Create(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
