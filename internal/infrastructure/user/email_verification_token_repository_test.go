package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"gorm.io/gorm"
)

func newTestUser(t *testing.T, tx *gorm.DB, email string) user.User {
	t.Helper()
	u := user.User{
		Username:       user.Username("Test"),
		Email:          user.Email(email),
		HashedPassword: user.HashedPassword("testpassword"),
		Role:           user.RoleUser,
	}
	if err := NewRepository(tx).Create(context.Background(), &u); err != nil {
		t.Fatalf(
			"Create(ctx, %v) error = %v",
			u, err,
		)
	}
	return u
}

func TestEmailVerificationTokenCreate(t *testing.T) {
	t.Run(
		"returns the token and stores only its hash when a valid user ID is given",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			evtr := NewEmailVerificationTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			evt, err := evtr.Create(ctx, u.ID)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v",
					u.ID, evt, err,
				)
			}

			sum := sha256.Sum256([]byte(evt.Value))
			hash := hex.EncodeToString(sum[:])
			if evt.UserID != u.ID ||
				evt.Value == "" ||
				string(evt.Hash) != hash ||
				string(evt.Hash) == string(evt.Value) {
				t.Errorf(
					"Create(ctx, %d) (*user.EmailVerificationToken, error) = %v, want UserID %v and the SHA-256 hash of the value",
					u.ID, evt, u.ID,
				)
			}
			if !evt.ExpiresAt.After(time.Now().Add(59 * time.Minute)) {
				t.Errorf(
					"Create(ctx, %d) ExpiresAt = %v, want about 1 hour later",
					u.ID, evt.ExpiresAt,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			evtr := NewEmailVerificationTokenRepository(tx, time.Hour)

			ctx := context.Background()
			fakeID := user.ID(999999)

			evt, err := evtr.Create(ctx, fakeID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Create(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v, want %v",
					fakeID, evt, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestEmailVerificationTokenFindValidByValue(t *testing.T) {
	t.Run(
		"returns the token when a valid value is given",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			evtr := NewEmailVerificationTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			want, err := evtr.Create(ctx, u.ID)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v",
					u.ID, want, err,
				)
			}

			got, err := evtr.FindValidByValue(ctx, want.Value)
			if err != nil {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*user.EmailVerificationToken, error) = %v, %v",
					want.Value, got, err,
				)
			}
			if got.ID != want.ID || got.UserID != want.UserID || got.Hash != want.Hash {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.EmailVerificationToken, error) = %v, want %v",
					want.Value, got, want,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the value does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			evtr := NewEmailVerificationTokenRepository(tx, time.Hour)

			ctx := context.Background()
			value := user.EmailVerificationTokenValue("notexist")

			got, err := evtr.FindValidByValue(ctx, value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*user.EmailVerificationToken, error) = %v, %v, want %v",
					value, got, err, exception.ErrInvalid,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the token is expired",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			evtr := NewEmailVerificationTokenRepository(tx, -time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			evt, err := evtr.Create(ctx, u.ID)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v",
					u.ID, evt, err,
				)
			}

			got, err := evtr.FindValidByValue(ctx, evt.Value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*user.EmailVerificationToken, error) = %v, %v, want %v",
					evt.Value, got, err, exception.ErrInvalid,
				)
			}
		},
	)
}

func TestEmailVerificationTokenFindLatestByUserID(t *testing.T) {
	t.Run(
		"returns the latest token of the user",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			evtr := NewEmailVerificationTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			first, err := evtr.Create(ctx, u.ID)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v",
					u.ID, first, err,
				)
			}
			latest, err := evtr.Create(ctx, u.ID)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v",
					u.ID, latest, err,
				)
			}

			got, err := evtr.FindLatestByUserID(ctx, u.ID)
			if err != nil {
				t.Fatalf(
					"FindLatestByUserID(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v",
					u.ID, got, err,
				)
			}
			if got.ID != latest.ID {
				t.Errorf(
					"FindLatestByUserID(ctx, %d) (*user.EmailVerificationToken, error) = %v, want %v",
					u.ID, got, latest,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user has no tokens",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			evtr := NewEmailVerificationTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			got, err := evtr.FindLatestByUserID(ctx, u.ID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"FindLatestByUserID(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v, want %v",
					u.ID, got, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestEmailVerificationTokenDeleteAllForUser(t *testing.T) {
	t.Run(
		"deletes all tokens of the user and keeps the tokens of other users",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			evtr := NewEmailVerificationTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u1 := newTestUser(t, tx, "test1@example.com")
			u2 := newTestUser(t, tx, "test2@example.com")

			evt1, err := evtr.Create(ctx, u1.ID)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v",
					u1.ID, evt1, err,
				)
			}
			evt2, err := evtr.Create(ctx, u2.ID)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d) (*user.EmailVerificationToken, error) = %v, %v",
					u2.ID, evt2, err,
				)
			}

			if err := evtr.DeleteAllForUser(ctx, u1.ID); err != nil {
				t.Fatalf(
					"DeleteAllForUser(ctx, %d) error = %v",
					u1.ID, err,
				)
			}

			if got, err := evtr.FindValidByValue(ctx, evt1.Value); !errors.Is(err, exception.ErrInvalid) {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.EmailVerificationToken, error) = %v, %v, want %v",
					evt1.Value, got, err, exception.ErrInvalid,
				)
			}
			if got, err := evtr.FindValidByValue(ctx, evt2.Value); err != nil {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.EmailVerificationToken, error) = %v, %v, want the token of the other user to remain",
					evt2.Value, got, err,
				)
			}
		},
	)
}
