package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepo struct {
	user *user.User
	err  error
}

func (r *fakeUserRepo) GetByID(
	ctx context.Context,
	userID user.ID,
) (*user.User, error) {
	return r.user, r.err
}

func (r *fakeUserRepo) GetByEmail(
	ctx context.Context,
	email user.Email,
) (*user.User, error) {
	return r.user, r.err
}

func (r *fakeUserRepo) Create(ctx context.Context, u *user.User) error {
	return nil
}

func (r *fakeUserRepo) ReplaceUnverified(ctx context.Context, u *user.User) error {
	return nil
}

func (r *fakeUserRepo) UpdateUsername(ctx context.Context, u *user.User) error {
	return nil
}

func (r *fakeUserRepo) UpdateEmail(ctx context.Context, u *user.User) error {
	return nil
}

func (r *fakeUserRepo) UpdatePassword(ctx context.Context, u *user.User) error {
	return nil
}

func (r *fakeUserRepo) UpdateEmailVerifiedAt(ctx context.Context, u *user.User) error {
	return nil
}

func (r *fakeUserRepo) Delete(ctx context.Context, userID user.ID) error {
	return nil
}

func (r *fakeUserRepo) DeleteUnverifiedCreatedBefore(ctx context.Context, before time.Time) error {
	return nil
}

type fakeAccessTokenService struct {
	principal *auth.Principal
}

func (r *fakeAccessTokenService) Generate(
	ctx context.Context,
	principal *auth.Principal,
) (*auth.AccessToken, error) {
	r.principal = principal
	return nil, nil
}

func (r *fakeAccessTokenService) Validate(
	ctx context.Context,
	accessToken *auth.AccessToken,
) (*auth.Principal, error) {
	return nil, nil
}

type fakeRefreshTokenRepo struct {
	refreshToken *auth.RefreshToken
	rotatedID    auth.RefreshTokenID
	rotateErr    error
	created      bool
	expiresAt    *time.Time
}

func (r *fakeRefreshTokenRepo) Create(
	ctx context.Context,
	principal *auth.Principal,
) (*auth.RefreshToken, error) {
	r.created = true
	return nil, nil
}

func (r *fakeRefreshTokenRepo) CreateWithExpiresAt(
	ctx context.Context,
	principal *auth.Principal,
	expiresAt time.Time,
) (*auth.RefreshToken, error) {
	r.expiresAt = &expiresAt
	return nil, nil
}

func (r *fakeRefreshTokenRepo) FindValidByValue(
	ctx context.Context,
	value auth.RefreshTokenValue,
) (*auth.RefreshToken, error) {
	return r.refreshToken, nil
}

func (r *fakeRefreshTokenRepo) FindRefreshableByValue(
	ctx context.Context,
	value auth.RefreshTokenValue,
) (*auth.RefreshToken, error) {
	return r.refreshToken, nil
}

func (r *fakeRefreshTokenRepo) Revoke(
	ctx context.Context,
	id auth.RefreshTokenID,
) error {
	return nil
}

func (r *fakeRefreshTokenRepo) Rotate(
	ctx context.Context,
	id auth.RefreshTokenID,
) error {
	r.rotatedID = id
	return r.rotateErr
}

func (r *fakeRefreshTokenRepo) RevokeAllForUser(
	ctx context.Context,
	userID user.ID,
) error {
	return nil
}

func (r *fakeRefreshTokenRepo) DeleteExpired(
	ctx context.Context,
) error {
	return nil
}

func TestLogin(t *testing.T) {
	t.Run(
		"builds the principal from the user when the credentials are valid",
		func(t *testing.T) {
			userID := user.ID(1)
			email := user.Email("test@example.com")
			password := user.Password("testpassword")
			hashed, err := bcrypt.GenerateFromPassword(
				[]byte(password),
				bcrypt.DefaultCost,
			)
			if err != nil {
				t.Fatalf(
					"GenerateFromPassword(%v, %v) = %v",
					password, bcrypt.DefaultCost, err,
				)
			}
			hashedPassword := user.HashedPassword(hashed)
			role := user.RoleUser
			emailVerifiedAt := time.Now()
			u := user.User{
				ID:              userID,
				Role:            role,
				HashedPassword:  hashedPassword,
				EmailVerifiedAt: &emailVerifiedAt,
			}

			userRepo := &fakeUserRepo{user: &u}
			accessTokenService := &fakeAccessTokenService{}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			au := NewAuthUsecase(userRepo, accessTokenService, refreshTokenRepo)

			ctx := context.Background()

			_, _, err = au.Login(ctx, email, password)
			if err != nil {
				t.Fatalf(
					"Login(ctx, %v, %v) (*auth.AccessToken, *auth.RefreshToken, error) = %v",
					email, password, err,
				)
			}
			principal := accessTokenService.principal
			if principal.UserID != userID || principal.Role != role {
				t.Errorf(
					"Login(ctx, %v, %v) Principal.UserID = %v, want %v, Principal.Role = %v, want %v",
					email, password, principal.UserID, userID, principal.Role, role,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the password does not match",
		func(t *testing.T) {
			email := user.Email("test@example.com")
			password := user.Password("testpassword")
			fakePassword := user.Password("fakepassword")
			hashed, err := bcrypt.GenerateFromPassword(
				[]byte(password),
				bcrypt.DefaultCost,
			)
			if err != nil {
				t.Fatalf(
					"GenerateFromPassword(%v, %v) = %v",
					password, bcrypt.DefaultCost, err,
				)
			}
			hashedPassword := user.HashedPassword(hashed)
			u := user.User{
				HashedPassword: hashedPassword,
			}

			userRepo := &fakeUserRepo{user: &u}
			accessTokenService := &fakeAccessTokenService{}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			au := NewAuthUsecase(userRepo, accessTokenService, refreshTokenRepo)

			ctx := context.Background()

			_, _, err = au.Login(ctx, email, fakePassword)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"Login(ctx, %v, %v) err = %v, want %v",
					email, fakePassword, err, exception.ErrInvalid,
				)
			}
		},
	)

	t.Run(
		"returns ErrForbidden without issuing tokens when the email is not verified",
		func(t *testing.T) {
			email := user.Email("test@example.com")
			password := user.Password("testpassword")
			hashed, err := bcrypt.GenerateFromPassword(
				[]byte(password),
				bcrypt.DefaultCost,
			)
			if err != nil {
				t.Fatalf(
					"GenerateFromPassword(%v, %v) = %v",
					password, bcrypt.DefaultCost, err,
				)
			}
			u := user.User{
				ID:             user.ID(1),
				Role:           user.RoleUser,
				HashedPassword: user.HashedPassword(hashed),
			}

			userRepo := &fakeUserRepo{user: &u}
			accessTokenService := &fakeAccessTokenService{}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			au := NewAuthUsecase(userRepo, accessTokenService, refreshTokenRepo)

			ctx := context.Background()

			_, _, err = au.Login(ctx, email, password)
			if !errors.Is(err, exception.ErrForbidden) {
				t.Fatalf(
					"Login(ctx, %v, %v) err = %v, want %v",
					email, password, err, exception.ErrForbidden,
				)
			}
			if accessTokenService.principal != nil {
				t.Errorf(
					"Login(ctx, %v, %v) generated an access token for %v, want none",
					email, password, accessTokenService.principal,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the user does not exist",
		func(t *testing.T) {
			email := user.Email("test@example.com")
			password := user.Password("testpassword")

			userRepo := &fakeUserRepo{err: exception.ErrNotFound}
			accessTokenService := &fakeAccessTokenService{}
			refreshTokenRepo := &fakeRefreshTokenRepo{}
			au := NewAuthUsecase(userRepo, accessTokenService, refreshTokenRepo)

			ctx := context.Background()

			_, _, err := au.Login(ctx, email, password)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"Login(ctx, %v, %v) err = %v, want %v",
					email, password, err, exception.ErrInvalid,
				)
			}
			if errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Login(ctx, %v, %v) err = %v, want an error that does not wrap %v",
					email, password, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestRefresh(t *testing.T) {
	t.Run(
		"rotates the old refresh token and builds the principal from the user when a valid value is given",
		func(t *testing.T) {
			userID := user.ID(1)
			role := user.RoleUser
			u := user.User{
				ID:   userID,
				Role: role,
			}

			oldRefreshTokenID := auth.RefreshTokenID(1)
			oldRefreshTokenValue := auth.RefreshTokenValue("refreshtokenvalue")
			oldPrincipal := auth.Principal{UserID: user.ID(2), Role: user.RoleAdmin}
			oldRefreshToken := auth.RefreshToken{
				ID:        oldRefreshTokenID,
				Value:     oldRefreshTokenValue,
				Principal: oldPrincipal,
			}

			userRepo := &fakeUserRepo{user: &u}
			accessTokenService := &fakeAccessTokenService{}
			refreshTokenRepo := &fakeRefreshTokenRepo{
				refreshToken: &oldRefreshToken,
			}
			au := NewAuthUsecase(userRepo, accessTokenService, refreshTokenRepo)

			ctx := context.Background()

			_, _, err := au.Refresh(ctx, oldRefreshTokenValue)
			if err != nil {
				t.Fatalf(
					"Refresh(ctx, %v) (*auth.AccessToken, *auth.RefreshToken, error) = %v",
					oldRefreshTokenValue, err,
				)
			}

			principal := accessTokenService.principal
			if principal.UserID != userID ||
				principal.Role != role {
				t.Errorf(
					"Refresh(ctx, %v) Principal.UserID = %v, want %v, Principal.Role = %v, want %v",
					oldRefreshTokenValue, principal.UserID, userID, principal.Role, role,
				)
			}

			if refreshTokenRepo.rotatedID != oldRefreshTokenID {
				t.Errorf(
					"Refresh(ctx, %v) rotated ID = %v, want %v",
					oldRefreshTokenValue, refreshTokenRepo.rotatedID, oldRefreshTokenID,
				)
			}
		},
	)

	t.Run(
		"issues a refresh token with a new expiry when the old one has not been rotated",
		func(t *testing.T) {
			u := user.User{ID: user.ID(1), Role: user.RoleUser}
			oldRefreshTokenValue := auth.RefreshTokenValue("refreshtokenvalue")
			oldRefreshToken := auth.RefreshToken{
				ID:        auth.RefreshTokenID(1),
				ExpiresAt: time.Now().Add(time.Hour),
			}

			userRepo := &fakeUserRepo{user: &u}
			accessTokenService := &fakeAccessTokenService{}
			refreshTokenRepo := &fakeRefreshTokenRepo{
				refreshToken: &oldRefreshToken,
			}
			au := NewAuthUsecase(userRepo, accessTokenService, refreshTokenRepo)

			ctx := context.Background()

			_, _, err := au.Refresh(ctx, oldRefreshTokenValue)
			if err != nil {
				t.Fatalf(
					"Refresh(ctx, %v) (*auth.AccessToken, *auth.RefreshToken, error) = %v",
					oldRefreshTokenValue, err,
				)
			}
			if !refreshTokenRepo.created || refreshTokenRepo.expiresAt != nil {
				t.Errorf(
					"Refresh(ctx, %v) created = %v, expiresAt = %v, want Create to be called",
					oldRefreshTokenValue, refreshTokenRepo.created, refreshTokenRepo.expiresAt,
				)
			}
		},
	)

	t.Run(
		"issues a refresh token with the same expiry when the old one has been rotated",
		func(t *testing.T) {
			u := user.User{ID: user.ID(1), Role: user.RoleUser}
			oldRefreshTokenValue := auth.RefreshTokenValue("refreshtokenvalue")
			rotatedAt := time.Now()
			oldExpiresAt := time.Now().Add(time.Hour)
			oldRefreshToken := auth.RefreshToken{
				ID:        auth.RefreshTokenID(1),
				ExpiresAt: oldExpiresAt,
				RotatedAt: &rotatedAt,
			}

			userRepo := &fakeUserRepo{user: &u}
			accessTokenService := &fakeAccessTokenService{}
			refreshTokenRepo := &fakeRefreshTokenRepo{
				refreshToken: &oldRefreshToken,
			}
			au := NewAuthUsecase(userRepo, accessTokenService, refreshTokenRepo)

			ctx := context.Background()

			_, _, err := au.Refresh(ctx, oldRefreshTokenValue)
			if err != nil {
				t.Fatalf(
					"Refresh(ctx, %v) (*auth.AccessToken, *auth.RefreshToken, error) = %v",
					oldRefreshTokenValue, err,
				)
			}
			if refreshTokenRepo.created ||
				refreshTokenRepo.expiresAt == nil ||
				!refreshTokenRepo.expiresAt.Equal(oldExpiresAt) {
				t.Errorf(
					"Refresh(ctx, %v) created = %v, expiresAt = %v, want %v",
					oldRefreshTokenValue, refreshTokenRepo.created, refreshTokenRepo.expiresAt, oldExpiresAt,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid without issuing tokens when the old refresh token cannot be rotated",
		func(t *testing.T) {
			u := user.User{ID: user.ID(1), Role: user.RoleUser}
			oldRefreshTokenValue := auth.RefreshTokenValue("refreshtokenvalue")
			oldRefreshToken := auth.RefreshToken{ID: auth.RefreshTokenID(1)}

			userRepo := &fakeUserRepo{user: &u}
			accessTokenService := &fakeAccessTokenService{}
			refreshTokenRepo := &fakeRefreshTokenRepo{
				refreshToken: &oldRefreshToken,
				rotateErr:    exception.ErrInvalid,
			}
			au := NewAuthUsecase(userRepo, accessTokenService, refreshTokenRepo)

			ctx := context.Background()

			_, _, err := au.Refresh(ctx, oldRefreshTokenValue)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"Refresh(ctx, %v) err = %v, want %v",
					oldRefreshTokenValue, err, exception.ErrInvalid,
				)
			}
			if accessTokenService.principal != nil ||
				refreshTokenRepo.created ||
				refreshTokenRepo.expiresAt != nil {
				t.Errorf(
					"Refresh(ctx, %v) issued tokens, want none",
					oldRefreshTokenValue,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the user does not exist",
		func(t *testing.T) {
			oldRefreshTokenValue := auth.RefreshTokenValue("refreshtokenvalue")
			oldRefreshToken := auth.RefreshToken{}

			userRepo := &fakeUserRepo{err: exception.ErrNotFound}
			accessTokenService := &fakeAccessTokenService{}
			refreshTokenRepo := &fakeRefreshTokenRepo{
				refreshToken: &oldRefreshToken,
			}
			au := NewAuthUsecase(userRepo, accessTokenService, refreshTokenRepo)

			ctx := context.Background()

			_, _, err := au.Refresh(ctx, oldRefreshTokenValue)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"Refresh(ctx, %v) err = %v, want %v",
					oldRefreshTokenValue, err, exception.ErrInvalid,
				)
			}
			if errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Refresh(ctx, %v) err = %v, want an error that does not wrap %v",
					oldRefreshTokenValue, err, exception.ErrNotFound,
				)
			}
		},
	)
}
