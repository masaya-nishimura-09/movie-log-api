package user

import (
	"context"
	"errors"
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
)

func newTestRepo(t *testing.T) user.UserRepository {
	t.Helper()
	db := testutil.TestDB(t)
	if err := db.Exec("TRUNCATE TABLE users RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to reset users table: %v", err)
	}
	return NewUserRepo(db)
}

func newTestUser(email string) *user.User {
	return &user.User{
		Username:       user.Username("Test User"),
		Email:          user.Email(email),
		HashedPassword: user.HashedPassword("hashed-password"),
		Role:           user.RoleUser,
	}
}

func TestUserRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	t.Run("creates a new user", func(t *testing.T) {
		u := newTestUser("create@example.com")
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		if u.ID == 0 {
			t.Error("Create() did not assign an ID")
		}
	})

	t.Run("rejects a duplicate email", func(t *testing.T) {
		if err := repo.Create(ctx, newTestUser("dup@example.com")); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		err := repo.Create(ctx, newTestUser("dup@example.com"))
		if !errors.Is(err, exception.ErrAlreadyExists) {
			t.Fatalf("Create() error = %v, want %v", err, exception.ErrAlreadyExists)
		}
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	t.Run("finds an existing user", func(t *testing.T) {
		u := newTestUser("getbyid@example.com")
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repo.GetByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got.Email != u.Email || got.Username != u.Username {
			t.Errorf("GetByID() = %+v, want email/username matching %+v", got, u)
		}
	})

	t.Run("returns not found for a missing user", func(t *testing.T) {
		_, err := repo.GetByID(ctx, user.ID(999999))
		if !errors.Is(err, exception.ErrNotFound) {
			t.Fatalf("GetByID() error = %v, want %v", err, exception.ErrNotFound)
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	t.Run("finds an existing user", func(t *testing.T) {
		u := newTestUser("getbyemail@example.com")
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repo.GetByEmail(ctx, u.Email)
		if err != nil {
			t.Fatalf("GetByEmail() error = %v", err)
		}
		if got.ID != u.ID {
			t.Errorf("GetByEmail() ID = %v, want %v", got.ID, u.ID)
		}
	})

	t.Run("returns not found for a missing email", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, user.Email("missing@example.com"))
		if !errors.Is(err, exception.ErrNotFound) {
			t.Fatalf("GetByEmail() error = %v, want %v", err, exception.ErrNotFound)
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	t.Run("updates an existing user", func(t *testing.T) {
		u := newTestUser("update@example.com")
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		u.Username = user.Username("Updated Name")
		u.Email = user.Email("updated@example.com")
		u.HashedPassword = user.HashedPassword("new-hashed-password")
		if err := repo.Update(ctx, u); err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		got, err := repo.GetByID(ctx, u.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if got.Username != "Updated Name" || got.Email != "updated@example.com" || got.HashedPassword != "new-hashed-password" {
			t.Errorf("Update() did not persist changes, got = %+v", got)
		}
	})

	t.Run("returns not found for a missing user", func(t *testing.T) {
		u := newTestUser("missing-update@example.com")
		u.ID = user.ID(999999)

		err := repo.Update(ctx, u)
		if !errors.Is(err, exception.ErrNotFound) {
			t.Fatalf("Update() error = %v, want %v", err, exception.ErrNotFound)
		}
	})
}

func TestUserRepository_Delete(t *testing.T) {
	ctx := context.Background()
	repo := newTestRepo(t)

	t.Run("deletes an existing user", func(t *testing.T) {
		u := newTestUser("delete@example.com")
		if err := repo.Create(ctx, u); err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if err := repo.Delete(ctx, u.ID); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err := repo.GetByID(ctx, u.ID)
		if !errors.Is(err, exception.ErrNotFound) {
			t.Fatalf("GetByID() after delete error = %v, want %v", err, exception.ErrNotFound)
		}
	})

	t.Run("returns not found for a missing user", func(t *testing.T) {
		err := repo.Delete(ctx, user.ID(999999))
		if !errors.Is(err, exception.ErrNotFound) {
			t.Fatalf("Delete() error = %v, want %v", err, exception.ErrNotFound)
		}
	})
}
