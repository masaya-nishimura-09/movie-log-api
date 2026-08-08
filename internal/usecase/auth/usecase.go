package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo  user.UserRepository
	accessTokenService auth.AccessTokenService
	refreshTokenRepo auth.RefreshTokenRepository
}


func NewAuthUsecase(
	userRepo user.UserRepository, 
	accessTokenService auth.AccessTokenService,
	refreshTokenRepo auth.RefreshTokenRepository,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo, 
		accessTokenService: accessTokenService,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (au *AuthUsecase) ValidateAccessToken(
	ctx context.Context, 
	accessToken *auth.AccessToken,
) (*auth.Principal, error) {
	return au.accessTokenService.Validate(ctx, accessToken)
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

	accessToken, err := au.accessTokenService.Generate(ctx, &principal)
	if err != nil {
		return nil, nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := au.refreshTokenRepo.Create(ctx, &principal)
	if err != nil {
		return nil, nil, fmt.Errorf("create refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}
