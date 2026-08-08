package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"golang.org/x/crypto/bcrypt"
)

func hashPassword(t *testing.T, password string) user.HashedPassword {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return user.HashedPassword(hashed)
}

func TestAuthUsecase_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("succeeds with correct credentials", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		userRepo.Users[1] = user.User{ID: 1, Email: "test@example.com", HashedPassword: hashPassword(t, "password1"), Role: user.RoleUser}
		uc := NewAuthUsecase(userRepo, &testutil.FakeAccessTokenService{}, testutil.NewFakeRefreshTokenRepository())

		accessToken, refreshToken, err := uc.Login(ctx, "test@example.com", "password1")
		if err != nil {
			t.Fatalf("Login() error = %v", err)
		}
		if accessToken.Value == "" || refreshToken.Value == "" {
			t.Error("Login() returned an empty token")
		}
	})

	t.Run("rejects an incorrect password", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		userRepo.Users[1] = user.User{ID: 1, Email: "test@example.com", HashedPassword: hashPassword(t, "password1"), Role: user.RoleUser}
		uc := NewAuthUsecase(userRepo, &testutil.FakeAccessTokenService{}, testutil.NewFakeRefreshTokenRepository())

		_, _, err := uc.Login(ctx, "test@example.com", "wrong-password")
		if !errors.Is(err, exception.ErrInvalid) {
			t.Fatalf("Login() error = %v, want %v", err, exception.ErrInvalid)
		}
	})

	t.Run("rejects an unknown email", func(t *testing.T) {
		uc := NewAuthUsecase(testutil.NewFakeUserRepository(), &testutil.FakeAccessTokenService{}, testutil.NewFakeRefreshTokenRepository())

		_, _, err := uc.Login(ctx, "missing@example.com", "password1")
		if !errors.Is(err, exception.ErrInvalid) {
			t.Fatalf("Login() error = %v, want %v", err, exception.ErrInvalid)
		}
	})
}

func TestAuthUsecase_Logout(t *testing.T) {
	ctx := context.Background()

	t.Run("revokes the refresh token", func(t *testing.T) {
		refreshRepo := testutil.NewFakeRefreshTokenRepository()
		created, err := refreshRepo.Create(ctx, &auth.Principal{UserID: 1, Role: user.RoleUser})
		if err != nil {
			t.Fatalf("failed to seed refresh token: %v", err)
		}
		uc := NewAuthUsecase(testutil.NewFakeUserRepository(), &testutil.FakeAccessTokenService{}, refreshRepo)

		if err := uc.Logout(ctx, created.Value); err != nil {
			t.Fatalf("Logout() error = %v", err)
		}
		if _, err := refreshRepo.FindValidByValue(ctx, created.Value); !errors.Is(err, exception.ErrInvalid) {
			t.Errorf("token still valid after Logout(), FindValidByValue() error = %v, want %v", err, exception.ErrInvalid)
		}
	})

	t.Run("propagates an unknown token", func(t *testing.T) {
		uc := NewAuthUsecase(testutil.NewFakeUserRepository(), &testutil.FakeAccessTokenService{}, testutil.NewFakeRefreshTokenRepository())

		err := uc.Logout(ctx, "does-not-exist")
		if !errors.Is(err, exception.ErrInvalid) {
			t.Fatalf("Logout() error = %v, want %v", err, exception.ErrInvalid)
		}
	})
}

func TestAuthUsecase_Refresh(t *testing.T) {
	ctx := context.Background()

	t.Run("rotates the token and uses the current user role", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		refreshRepo := testutil.NewFakeRefreshTokenRepository()
		created, err := refreshRepo.Create(ctx, &auth.Principal{UserID: 1, Role: user.RoleUser})
		if err != nil {
			t.Fatalf("failed to seed refresh token: %v", err)
		}
		userRepo.Users[1] = user.User{ID: 1, Role: user.RoleAdmin}

		uc := NewAuthUsecase(userRepo, &testutil.FakeAccessTokenService{}, refreshRepo)

		_, newRefreshToken, err := uc.Refresh(ctx, created.Value)
		if err != nil {
			t.Fatalf("Refresh() error = %v", err)
		}
		if newRefreshToken.Principal.Role != user.RoleAdmin {
			t.Errorf("Refresh() Role = %v, want %v (the current role, not the stale one)", newRefreshToken.Principal.Role, user.RoleAdmin)
		}
		if _, err := refreshRepo.FindValidByValue(ctx, created.Value); !errors.Is(err, exception.ErrInvalid) {
			t.Errorf("old token still valid after Refresh(), FindValidByValue() error = %v, want %v", err, exception.ErrInvalid)
		}
	})

	t.Run("rejects a token for a deleted user", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		refreshRepo := testutil.NewFakeRefreshTokenRepository()
		created, err := refreshRepo.Create(ctx, &auth.Principal{UserID: 1, Role: user.RoleUser})
		if err != nil {
			t.Fatalf("failed to seed refresh token: %v", err)
		}
		uc := NewAuthUsecase(userRepo, &testutil.FakeAccessTokenService{}, refreshRepo)

		_, _, err = uc.Refresh(ctx, created.Value)
		if !errors.Is(err, exception.ErrInvalid) {
			t.Fatalf("Refresh() error = %v, want %v", err, exception.ErrInvalid)
		}
	})

	t.Run("propagates an unknown token", func(t *testing.T) {
		uc := NewAuthUsecase(testutil.NewFakeUserRepository(), &testutil.FakeAccessTokenService{}, testutil.NewFakeRefreshTokenRepository())

		_, _, err := uc.Refresh(ctx, "does-not-exist")
		if !errors.Is(err, exception.ErrInvalid) {
			t.Fatalf("Refresh() error = %v, want %v", err, exception.ErrInvalid)
		}
	})
}

func TestAuthUsecase_ValidateAccessToken(t *testing.T) {
	ctx := context.Background()
	wantPrincipal := &auth.Principal{UserID: 1, Role: user.RoleAdmin}
	tokenService := &testutil.FakeAccessTokenService{ValidatePrincipal: wantPrincipal}
	uc := NewAuthUsecase(testutil.NewFakeUserRepository(), tokenService, testutil.NewFakeRefreshTokenRepository())

	got, err := uc.ValidateAccessToken(ctx, &auth.AccessToken{Value: "some-token"})
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if got != wantPrincipal {
		t.Errorf("ValidateAccessToken() = %v, want %v", got, wantPrincipal)
	}
}
