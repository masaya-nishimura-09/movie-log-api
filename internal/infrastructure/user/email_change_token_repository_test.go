package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/mail"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
)

func TestEmailChangeTokenCreate(t *testing.T) {
	t.Run(
		"returns the token and stores only its hash when a valid user ID is given",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			ectr := NewEmailChangeTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")
			newEmail := user.Email("new@example.com")

			ect, err := ectr.Create(ctx, u.ID, newEmail, mail.LanguageEN)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v, %v) (*user.EmailChangeToken, error) = %v, %v",
					u.ID, newEmail, mail.LanguageEN, ect, err,
				)
			}

			sum := sha256.Sum256([]byte(ect.Value))
			hash := hex.EncodeToString(sum[:])
			if ect.UserID != u.ID ||
				ect.NewEmail != newEmail ||
				ect.Language != mail.LanguageEN ||
				ect.Value == "" ||
				string(ect.Hash) != hash {
				t.Errorf(
					"Create(ctx, %d, %v, %v) (*user.EmailChangeToken, error) = %v, want UserID %v, NewEmail %v, Language %v and the SHA-256 hash of the value",
					u.ID, newEmail, mail.LanguageEN, ect, u.ID, newEmail, mail.LanguageEN,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			ectr := NewEmailChangeTokenRepository(tx, time.Hour)

			ctx := context.Background()
			fakeID := user.ID(999999)

			ect, err := ectr.Create(ctx, fakeID, user.Email("new@example.com"), mail.LanguageJA)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Create(ctx, %d, ...) (*user.EmailChangeToken, error) = %v, %v, want %v",
					fakeID, ect, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestEmailChangeTokenFindValidByValue(t *testing.T) {
	t.Run(
		"returns the token with the new email and language when a valid value is given",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			ectr := NewEmailChangeTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			want, err := ectr.Create(ctx, u.ID, user.Email("new@example.com"), mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, ...) (*user.EmailChangeToken, error) = %v, %v",
					u.ID, want, err,
				)
			}

			got, err := ectr.FindValidByValue(ctx, want.Value)
			if err != nil {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*user.EmailChangeToken, error) = %v, %v",
					want.Value, got, err,
				)
			}
			if got.ID != want.ID ||
				got.UserID != want.UserID ||
				got.NewEmail != want.NewEmail ||
				got.Language != want.Language ||
				got.Hash != want.Hash {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.EmailChangeToken, error) = %v, want %v",
					want.Value, got, want,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the value does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			ectr := NewEmailChangeTokenRepository(tx, time.Hour)

			ctx := context.Background()
			value := user.EmailChangeTokenValue("notexist")

			got, err := ectr.FindValidByValue(ctx, value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*user.EmailChangeToken, error) = %v, %v, want %v",
					value, got, err, exception.ErrInvalid,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the token is expired",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			ectr := NewEmailChangeTokenRepository(tx, -time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			ect, err := ectr.Create(ctx, u.ID, user.Email("new@example.com"), mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, ...) (*user.EmailChangeToken, error) = %v, %v",
					u.ID, ect, err,
				)
			}

			got, err := ectr.FindValidByValue(ctx, ect.Value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*user.EmailChangeToken, error) = %v, %v, want %v",
					ect.Value, got, err, exception.ErrInvalid,
				)
			}
		},
	)
}

func TestEmailChangeTokenDeleteAllForUser(t *testing.T) {
	t.Run(
		"deletes all tokens of the user and keeps the tokens of other users",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			ectr := NewEmailChangeTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u1 := newTestUser(t, tx, "test1@example.com")
			u2 := newTestUser(t, tx, "test2@example.com")

			ect1, err := ectr.Create(ctx, u1.ID, user.Email("new1@example.com"), mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, ...) (*user.EmailChangeToken, error) = %v, %v",
					u1.ID, ect1, err,
				)
			}
			ect2, err := ectr.Create(ctx, u2.ID, user.Email("new2@example.com"), mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, ...) (*user.EmailChangeToken, error) = %v, %v",
					u2.ID, ect2, err,
				)
			}

			if err := ectr.DeleteAllForUser(ctx, u1.ID); err != nil {
				t.Fatalf(
					"DeleteAllForUser(ctx, %d) error = %v",
					u1.ID, err,
				)
			}

			if got, err := ectr.FindValidByValue(ctx, ect1.Value); !errors.Is(err, exception.ErrInvalid) {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.EmailChangeToken, error) = %v, %v, want %v",
					ect1.Value, got, err, exception.ErrInvalid,
				)
			}
			if got, err := ectr.FindValidByValue(ctx, ect2.Value); err != nil {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.EmailChangeToken, error) = %v, %v, want the token of the other user to remain",
					ect2.Value, got, err,
				)
			}
		},
	)
}

func TestEmailChangeTokenDeleteExpired(t *testing.T) {
	t.Run(
		"deletes the expired tokens and keeps the valid one",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			ctx := context.Background()
			u1 := newTestUser(t, tx, "test1@example.com")
			u2 := newTestUser(t, tx, "test2@example.com")

			expired, err := NewEmailChangeTokenRepository(tx, -time.Hour).
				Create(ctx, u1.ID, user.Email("new1@example.com"), mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, ...) (*user.EmailChangeToken, error) = %v, %v",
					u1.ID, expired, err,
				)
			}

			ectr := NewEmailChangeTokenRepository(tx, time.Hour)
			valid, err := ectr.Create(ctx, u2.ID, user.Email("new2@example.com"), mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, ...) (*user.EmailChangeToken, error) = %v, %v",
					u2.ID, valid, err,
				)
			}

			if err := ectr.DeleteExpired(ctx); err != nil {
				t.Fatalf("DeleteExpired(ctx) error = %v", err)
			}

			var count int64
			if err := tx.Model(&emailChangeTokenDTO{}).Where("id = ?", uint(expired.ID)).Count(&count).Error; err != nil {
				t.Fatalf("Count() error = %v", err)
			}
			if count != 0 {
				t.Errorf("DeleteExpired(ctx) left the expired token %d, want it deleted", expired.ID)
			}
			if got, err := ectr.FindValidByValue(ctx, valid.Value); err != nil {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.EmailChangeToken, error) = %v, %v, want the valid token to remain",
					valid.Value, got, err,
				)
			}
		},
	)
}
