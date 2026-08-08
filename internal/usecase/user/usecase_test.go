package user

import (
	"context"
	"errors"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"golang.org/x/crypto/bcrypt"
)

func TestUserUsecase_GetByID(t *testing.T) {
	ctx := context.Background()
	userRepo := testutil.NewFakeUserRepository()
	uc := NewUserUsecase(userRepo, testutil.NewFakeRefreshTokenRepository())

	t.Run("returns an existing user", func(t *testing.T) {
		userRepo.Users[1] = user.User{ID: 1, Username: "Test", Email: "test@example.com"}

		got, err := uc.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got.ID != 1 {
			t.Errorf("GetByID() ID = %v, want 1", got.ID)
		}
	})

	t.Run("propagates not found", func(t *testing.T) {
		_, err := uc.GetByID(ctx, 999)
		if !errors.Is(err, exception.ErrNotFound) {
			t.Fatalf("GetByID() error = %v, want %v", err, exception.ErrNotFound)
		}
	})
}

func TestUserUsecase_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("hashes the password and creates the user", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		uc := NewUserUsecase(userRepo, testutil.NewFakeRefreshTokenRepository())

		got, err := uc.Register(ctx, "Test User", "test@example.com", "password1")
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if got.Role != user.RoleUser {
			t.Errorf("Register() Role = %v, want %v", got.Role, user.RoleUser)
		}
		if err := bcrypt.CompareHashAndPassword([]byte(got.HashedPassword), []byte("password1")); err != nil {
			t.Errorf("Register() password not correctly hashed: %v", err)
		}
	})

	t.Run("propagates a repository error", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		userRepo.CreateErr = exception.ErrAlreadyExists
		uc := NewUserUsecase(userRepo, testutil.NewFakeRefreshTokenRepository())

		_, err := uc.Register(ctx, "Test User", "test@example.com", "password1")
		if !errors.Is(err, exception.ErrAlreadyExists) {
			t.Fatalf("Register() error = %v, want %v", err, exception.ErrAlreadyExists)
		}
	})
}

func TestUserUsecase_UpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("updates an existing user", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		userRepo.Users[1] = user.User{ID: 1, Username: "Old", Email: "old@example.com"}
		uc := NewUserUsecase(userRepo, testutil.NewFakeRefreshTokenRepository())

		got, err := uc.UpdateUser(ctx, 1, "New", "new@example.com", "password1")
		if err != nil {
			t.Fatalf("UpdateUser() error = %v", err)
		}
		if got.Username != "New" || got.Email != "new@example.com" {
			t.Errorf("UpdateUser() = %+v, want updated username/email", got)
		}
	})

	t.Run("propagates not found", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		uc := NewUserUsecase(userRepo, testutil.NewFakeRefreshTokenRepository())

		_, err := uc.UpdateUser(ctx, 999, "New", "new@example.com", "password1")
		if !errors.Is(err, exception.ErrNotFound) {
			t.Fatalf("UpdateUser() error = %v, want %v", err, exception.ErrNotFound)
		}
	})
}

func TestUserUsecase_DeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes the user and revokes their refresh tokens", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		userRepo.Users[1] = user.User{ID: 1}
		refreshRepo := testutil.NewFakeRefreshTokenRepository()
		uc := NewUserUsecase(userRepo, refreshRepo)

		if err := uc.DeleteUser(ctx, 1); err != nil {
			t.Fatalf("DeleteUser() error = %v", err)
		}
		if len(refreshRepo.RevokeAllForUserCalls) != 1 || refreshRepo.RevokeAllForUserCalls[0] != 1 {
			t.Errorf("DeleteUser() RevokeAllForUser calls = %v, want [1]", refreshRepo.RevokeAllForUserCalls)
		}
	})

	t.Run("does not revoke tokens when delete fails", func(t *testing.T) {
		userRepo := testutil.NewFakeUserRepository()
		refreshRepo := testutil.NewFakeRefreshTokenRepository()
		uc := NewUserUsecase(userRepo, refreshRepo)

		err := uc.DeleteUser(ctx, 999)
		if !errors.Is(err, exception.ErrNotFound) {
			t.Fatalf("DeleteUser() error = %v, want %v", err, exception.ErrNotFound)
		}
		if len(refreshRepo.RevokeAllForUserCalls) != 0 {
			t.Errorf("DeleteUser() RevokeAllForUser calls = %v, want none", refreshRepo.RevokeAllForUserCalls)
		}
	})
}
