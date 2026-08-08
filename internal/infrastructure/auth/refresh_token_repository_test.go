package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
)

func newTestRefreshTokenRepo(t *testing.T, ttl time.Duration) auth.RefreshTokenRepository {
	t.Helper()
	db := testutil.TestDB(t)
	if err := db.Exec("TRUNCATE TABLE refresh_tokens RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to reset refresh_tokens table: %v", err)
	}
	return NewRefreshTokenRepo(db, ttl)
}

func testPrincipal(userID user.ID) *auth.Principal {
	return &auth.Principal{UserID: userID, Role: user.RoleUser}
}

func TestRefreshTokenRepository_Create(t *testing.T) {
	ctx := context.Background()
	repo := newTestRefreshTokenRepo(t, time.Hour)

	got, err := repo.Create(ctx, testPrincipal(1))
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.Value == "" {
		t.Error("Create() returned an empty token value")
	}
	if got.ID == 0 {
		t.Error("Create() did not assign an ID")
	}

	sum := sha256.Sum256([]byte(got.Value))
	wantHash := auth.RefreshTokenHash(hex.EncodeToString(sum[:]))
	if got.Hash != wantHash {
		t.Errorf("Create() Hash = %v, want %v (sha256 of the returned value)", got.Hash, wantHash)
	}

	if !got.ExpiresAt.After(time.Now()) {
		t.Errorf("Create() ExpiresAt = %v, want a time in the future", got.ExpiresAt)
	}
}

func TestRefreshTokenRepository_FindValidByValue(t *testing.T) {
	ctx := context.Background()

	t.Run("finds a valid token", func(t *testing.T) {
		repo := newTestRefreshTokenRepo(t, time.Hour)
		principal := testPrincipal(1)
		created, err := repo.Create(ctx, principal)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		found, err := repo.FindValidByValue(ctx, created.Value)
		if err != nil {
			t.Fatalf("FindValidByValue() error = %v", err)
		}
		if found.Principal.UserID != principal.UserID || found.Principal.Role != principal.Role {
			t.Errorf("FindValidByValue() Principal = %+v, want %+v", found.Principal, principal)
		}
	})

	t.Run("rejects an unknown value", func(t *testing.T) {
		repo := newTestRefreshTokenRepo(t, time.Hour)

		_, err := repo.FindValidByValue(ctx, auth.RefreshTokenValue("does-not-exist"))
		if !errors.Is(err, exception.ErrInvalid) {
			t.Fatalf("FindValidByValue() error = %v, want %v", err, exception.ErrInvalid)
		}
	})

	t.Run("rejects an expired token", func(t *testing.T) {
		repo := newTestRefreshTokenRepo(t, -time.Hour)
		created, err := repo.Create(ctx, testPrincipal(1))
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		_, err = repo.FindValidByValue(ctx, created.Value)
		if !errors.Is(err, exception.ErrInvalid) {
			t.Fatalf("FindValidByValue() error = %v, want %v", err, exception.ErrInvalid)
		}
	})

	t.Run("rejects a revoked token", func(t *testing.T) {
		repo := newTestRefreshTokenRepo(t, time.Hour)
		created, err := repo.Create(ctx, testPrincipal(1))
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		found, err := repo.FindValidByValue(ctx, created.Value)
		if err != nil {
			t.Fatalf("FindValidByValue() error = %v", err)
		}
		if err := repo.Revoke(ctx, found.ID); err != nil {
			t.Fatalf("Revoke() error = %v", err)
		}

		_, err = repo.FindValidByValue(ctx, created.Value)
		if !errors.Is(err, exception.ErrInvalid) {
			t.Fatalf("FindValidByValue() after revoke error = %v, want %v", err, exception.ErrInvalid)
		}
	})
}

func TestRefreshTokenRepository_Revoke(t *testing.T) {
	ctx := context.Background()

	t.Run("revokes an existing token", func(t *testing.T) {
		repo := newTestRefreshTokenRepo(t, time.Hour)
		created, err := repo.Create(ctx, testPrincipal(1))
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}
		found, err := repo.FindValidByValue(ctx, created.Value)
		if err != nil {
			t.Fatalf("FindValidByValue() error = %v", err)
		}

		if err := repo.Revoke(ctx, found.ID); err != nil {
			t.Fatalf("Revoke() error = %v", err)
		}
	})

	t.Run("rejects an unknown id", func(t *testing.T) {
		repo := newTestRefreshTokenRepo(t, time.Hour)

		err := repo.Revoke(ctx, auth.RefreshTokenID(999999))
		if !errors.Is(err, exception.ErrInvalid) {
			t.Fatalf("Revoke() error = %v, want %v", err, exception.ErrInvalid)
		}
	})
}

func TestRefreshTokenRepository_RevokeAllForUser(t *testing.T) {
	ctx := context.Background()
	repo := newTestRefreshTokenRepo(t, time.Hour)

	userA := testPrincipal(1)
	userB := testPrincipal(2)

	tokenA1, err := repo.Create(ctx, userA)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	tokenA2, err := repo.Create(ctx, userA)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	tokenB, err := repo.Create(ctx, userB)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.RevokeAllForUser(ctx, userA.UserID); err != nil {
		t.Fatalf("RevokeAllForUser() error = %v", err)
	}

	if _, err := repo.FindValidByValue(ctx, tokenA1.Value); !errors.Is(err, exception.ErrInvalid) {
		t.Errorf("token A1 FindValidByValue() error = %v, want %v", err, exception.ErrInvalid)
	}
	if _, err := repo.FindValidByValue(ctx, tokenA2.Value); !errors.Is(err, exception.ErrInvalid) {
		t.Errorf("token A2 FindValidByValue() error = %v, want %v", err, exception.ErrInvalid)
	}
	if _, err := repo.FindValidByValue(ctx, tokenB.Value); err != nil {
		t.Errorf("token B (different user) FindValidByValue() error = %v, want nil", err)
	}

	t.Run("no error when the user has no tokens", func(t *testing.T) {
		if err := repo.RevokeAllForUser(ctx, user.ID(999999)); err != nil {
			t.Fatalf("RevokeAllForUser() error = %v", err)
		}
	})
}
