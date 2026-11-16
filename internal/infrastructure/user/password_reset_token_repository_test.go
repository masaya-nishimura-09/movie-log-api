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

func TestPasswordResetTokenCreate(t *testing.T) {
	t.Run(
		"returns the token and stores only its hash when a valid user ID is given",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			prtr := NewPasswordResetTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			prt, err := prtr.Create(ctx, u.ID, mail.LanguageEN)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v",
					u.ID, mail.LanguageEN, prt, err,
				)
			}

			sum := sha256.Sum256([]byte(prt.Value))
			hash := hex.EncodeToString(sum[:])
			if prt.UserID != u.ID ||
				prt.Language != mail.LanguageEN ||
				prt.Value == "" ||
				string(prt.Hash) != hash {
				t.Errorf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, want UserID %v, Language %v and the SHA-256 hash of the value",
					u.ID, mail.LanguageEN, prt, u.ID, mail.LanguageEN,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			prtr := NewPasswordResetTokenRepository(tx, time.Hour)

			ctx := context.Background()
			fakeID := user.ID(999999)

			prt, err := prtr.Create(ctx, fakeID, mail.LanguageJA)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v, want %v",
					fakeID, mail.LanguageJA, prt, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestPasswordResetTokenFindValidByValue(t *testing.T) {
	t.Run(
		"returns the token when a valid value is given",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			prtr := NewPasswordResetTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			want, err := prtr.Create(ctx, u.ID, mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v",
					u.ID, mail.LanguageJA, want, err,
				)
			}

			got, err := prtr.FindValidByValue(ctx, want.Value)
			if err != nil {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*user.PasswordResetToken, error) = %v, %v",
					want.Value, got, err,
				)
			}
			if got.ID != want.ID ||
				got.UserID != want.UserID ||
				got.Language != want.Language ||
				got.Hash != want.Hash {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.PasswordResetToken, error) = %v, want %v",
					want.Value, got, want,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the value does not exist",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			prtr := NewPasswordResetTokenRepository(tx, time.Hour)

			ctx := context.Background()
			value := user.PasswordResetTokenValue("notexist")

			got, err := prtr.FindValidByValue(ctx, value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*user.PasswordResetToken, error) = %v, %v, want %v",
					value, got, err, exception.ErrInvalid,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the token is expired",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			prtr := NewPasswordResetTokenRepository(tx, -time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			prt, err := prtr.Create(ctx, u.ID, mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v",
					u.ID, mail.LanguageJA, prt, err,
				)
			}

			got, err := prtr.FindValidByValue(ctx, prt.Value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*user.PasswordResetToken, error) = %v, %v, want %v",
					prt.Value, got, err, exception.ErrInvalid,
				)
			}
		},
	)
}

func TestPasswordResetTokenFindLatestByUserID(t *testing.T) {
	t.Run(
		"returns the latest token of the user",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			prtr := NewPasswordResetTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			first, err := prtr.Create(ctx, u.ID, mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v",
					u.ID, mail.LanguageJA, first, err,
				)
			}
			latest, err := prtr.Create(ctx, u.ID, mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v",
					u.ID, mail.LanguageJA, latest, err,
				)
			}

			got, err := prtr.FindLatestByUserID(ctx, u.ID)
			if err != nil {
				t.Fatalf(
					"FindLatestByUserID(ctx, %d) (*user.PasswordResetToken, error) = %v, %v",
					u.ID, got, err,
				)
			}
			if got.ID != latest.ID {
				t.Errorf(
					"FindLatestByUserID(ctx, %d) (*user.PasswordResetToken, error) = %v, want %v",
					u.ID, got, latest,
				)
			}
		},
	)

	t.Run(
		"returns ErrNotFound when the user has no tokens",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			prtr := NewPasswordResetTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u := newTestUser(t, tx, "test@example.com")

			got, err := prtr.FindLatestByUserID(ctx, u.ID)
			if !errors.Is(err, exception.ErrNotFound) {
				t.Fatalf(
					"FindLatestByUserID(ctx, %d) (*user.PasswordResetToken, error) = %v, %v, want %v",
					u.ID, got, err, exception.ErrNotFound,
				)
			}
		},
	)
}

func TestPasswordResetTokenDeleteAllForUser(t *testing.T) {
	t.Run(
		"deletes all tokens of the user and keeps the tokens of other users",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			prtr := NewPasswordResetTokenRepository(tx, time.Hour)

			ctx := context.Background()
			u1 := newTestUser(t, tx, "test1@example.com")
			u2 := newTestUser(t, tx, "test2@example.com")

			prt1, err := prtr.Create(ctx, u1.ID, mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v",
					u1.ID, mail.LanguageJA, prt1, err,
				)
			}
			prt2, err := prtr.Create(ctx, u2.ID, mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v",
					u2.ID, mail.LanguageJA, prt2, err,
				)
			}

			if err := prtr.DeleteAllForUser(ctx, u1.ID); err != nil {
				t.Fatalf(
					"DeleteAllForUser(ctx, %d) error = %v",
					u1.ID, err,
				)
			}

			if got, err := prtr.FindValidByValue(ctx, prt1.Value); !errors.Is(err, exception.ErrInvalid) {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.PasswordResetToken, error) = %v, %v, want %v",
					prt1.Value, got, err, exception.ErrInvalid,
				)
			}
			if got, err := prtr.FindValidByValue(ctx, prt2.Value); err != nil {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.PasswordResetToken, error) = %v, %v, want the token of the other user to remain",
					prt2.Value, got, err,
				)
			}
		},
	)
}

func TestPasswordResetTokenDeleteExpired(t *testing.T) {
	t.Run(
		"deletes the expired tokens and keeps the valid one",
		func(t *testing.T) {
			tx := testutil.BeginTx(t, testDB)
			ctx := context.Background()
			u1 := newTestUser(t, tx, "test1@example.com")
			u2 := newTestUser(t, tx, "test2@example.com")

			expired, err := NewPasswordResetTokenRepository(tx, -time.Hour).
				Create(ctx, u1.ID, mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v",
					u1.ID, mail.LanguageJA, expired, err,
				)
			}

			prtr := NewPasswordResetTokenRepository(tx, time.Hour)
			valid, err := prtr.Create(ctx, u2.ID, mail.LanguageJA)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %d, %v) (*user.PasswordResetToken, error) = %v, %v",
					u2.ID, mail.LanguageJA, valid, err,
				)
			}

			if err := prtr.DeleteExpired(ctx); err != nil {
				t.Fatalf("DeleteExpired(ctx) error = %v", err)
			}

			var count int64
			if err := tx.Model(&passwordResetTokenDTO{}).Where("id = ?", uint(expired.ID)).Count(&count).Error; err != nil {
				t.Fatalf("Count() error = %v", err)
			}
			if count != 0 {
				t.Errorf("DeleteExpired(ctx) left the expired token %d, want it deleted", expired.ID)
			}
			if got, err := prtr.FindValidByValue(ctx, valid.Value); err != nil {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*user.PasswordResetToken, error) = %v, %v, want the valid token to remain",
					valid.Value, got, err,
				)
			}
		},
	)
}
