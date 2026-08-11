package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
	"github.com/masaya-nishimura-09/movie-log-api/internal/testutil"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	testDB = testutil.NewTestDB()
	code := m.Run()
	os.Exit(code)
}

func newTestRepo(t *testing.T, ttl time.Duration) auth.RefreshTokenRepository {
	t.Helper()
	return NewRefreshTokenRepo(testutil.BeginTx(t, testDB), ttl)
}

func TestCreate(t *testing.T) {
	t.Run(
		"returns the refresh token when valid principal is given",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx := context.Background()
			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			rt, err := rtr.Create(ctx, &principal)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
					principal, rt, err,
				)
			}

			sum := sha256.Sum256([]byte(rt.Value))
			hash := hex.EncodeToString(sum[:])
			if rt.Hash != auth.RefreshTokenHash(hash) {
				t.Errorf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, want hash %v",
					principal, rt, hash,
				)
			}

			if rt.Principal.UserID != principal.UserID ||
				rt.Principal.Role != principal.Role {
				t.Errorf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, want principal %v",
					principal, rt, principal,
				)
			}

			if rt.CreatedAt.IsZero() {
				t.Errorf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, want non-zero CreatedAt",
					principal, rt,
				)
			}

			found, err := rtr.FindValidByValue(ctx, rt.Value)
			if err != nil {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
					rt.Value, found, err,
				)
			}
			if found.ID != rt.ID {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, want ID %v",
					rt.Value, found, rt.ID,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			got, err := rtr.Create(ctx, &principal)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, %v, want %v",
					principal, got, err, context.Canceled,
				)
			}
			if got != nil {
				t.Errorf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, %v, want nil",
					principal, got, err,
				)
			}
		},
	)
}

func TestFindValidByValue(t *testing.T) {
	t.Run(
		"returns the refresh token when valid value is given",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx := context.Background()
			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			rt, err := rtr.Create(ctx, &principal)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
					principal, rt, err,
				)
			}

			found, err := rtr.FindValidByValue(ctx, rt.Value)
			if err != nil {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
					rt.Value, found, err,
				)
			}
			if rt.ID != found.ID ||
				rt.Principal.UserID != found.Principal.UserID ||
				rt.Principal.Role != found.Principal.Role ||
				rt.Hash != found.Hash ||
				!rt.CreatedAt.Truncate(time.Microsecond).Equal(found.CreatedAt) ||
				!rt.ExpiresAt.Truncate(time.Microsecond).Equal(found.ExpiresAt) ||
				found.RevokedAt != nil {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, want %v",
					rt.Value, found, rt,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when value does not exist",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx := context.Background()

			fakeValue := auth.RefreshTokenValue("")

			found, err := rtr.FindValidByValue(ctx, fakeValue)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, %v, want %v",
					fakeValue, found, err, exception.ErrInvalid,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when refresh token is expired",
		func(t *testing.T) {
			rtr := newTestRepo(t, -time.Hour)

			ctx := context.Background()

			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			rt, err := rtr.Create(ctx, &principal)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
					principal, rt, err,
				)
			}

			found, err := rtr.FindValidByValue(ctx, rt.Value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, %v, want %v",
					rt.Value, found, err, exception.ErrInvalid,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			value := auth.RefreshTokenValue("")

			got, err := rtr.FindValidByValue(ctx, value)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, %v, want %v",
					value, got, err, context.Canceled,
				)
			}
			if got != nil {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, %v, want nil",
					value, got, err,
				)
			}
		},
	)
}

func TestRevoke(t *testing.T) {
	t.Run(
		"revokes the refresh token when a valid id is given",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx := context.Background()
			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			rt, err := rtr.Create(ctx, &principal)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
					principal, rt, err,
				)
			}

			if err := rtr.Revoke(ctx, rt.ID); err != nil {
				t.Fatalf(
					"Revoke(ctx, %d) error = %v",
					rt.ID, err,
				)
			}

			found, err := rtr.FindValidByValue(ctx, rt.Value)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, %v, want %v",
					rt.Value, found, err, exception.ErrInvalid,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when the refresh token is already revoked",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx := context.Background()
			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			rt, err := rtr.Create(ctx, &principal)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
					principal, rt, err,
				)
			}

			if err := rtr.Revoke(ctx, rt.ID); err != nil {
				t.Fatalf(
					"Revoke(ctx, %d) error = %v",
					rt.ID, err,
				)
			}

			if err := rtr.Revoke(ctx, rt.ID); !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"Revoke(ctx, %d) error = %v, want %v",
					rt.ID, err, exception.ErrInvalid,
				)
			}
		},
	)

	t.Run(
		"returns ErrInvalid when refresh token is not found",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx := context.Background()
			id := auth.RefreshTokenID(1)

			err := rtr.Revoke(ctx, id)
			if !errors.Is(err, exception.ErrInvalid) {
				t.Fatalf(
					"Revoke(ctx, %d) error = %v, want %v",
					id, err, exception.ErrInvalid,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			id := auth.RefreshTokenID(1)

			err := rtr.Revoke(ctx, id)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"Revoke(ctx, %d) error = %v, want %v",
					id, err, context.Canceled,
				)
			}
		},
	)
}

func TestRevokeAllForUser(t *testing.T) {
	t.Run(
		"revokes the all refresh tokens when a valid user id is given",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx := context.Background()

			userID := user.ID(1)
			userID2 := user.ID(2)
			principal := auth.Principal{
				UserID: userID,
				Role:   user.RoleAdmin,
			}
			principal2 := auth.Principal{
				UserID: userID2,
				Role:   user.RoleAdmin,
			}
			var tokens []*auth.RefreshToken

			for range 5 {
				rt, err := rtr.Create(ctx, &principal)
				if err != nil {
					t.Fatalf(
						"Create(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
						principal, rt, err,
					)
				}
				tokens = append(tokens, rt)
			}

			otherToken, err := rtr.Create(ctx, &principal2)
			if err != nil {
				t.Fatalf(
					"Create(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
					principal2, otherToken, err,
				)
			}

			if err := rtr.RevokeAllForUser(ctx, userID); err != nil {
				t.Fatalf(
					"RevokeAllForUser(ctx, %d) error = %v",
					userID, err,
				)
			}

			for _, v := range tokens {
				found, err := rtr.FindValidByValue(ctx, v.Value)
				if !errors.Is(err, exception.ErrInvalid) {
					t.Errorf(
						"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, %v, want %v",
						v.Value, found, err, exception.ErrInvalid,
					)
				}
			}

			found, err := rtr.FindValidByValue(ctx, otherToken.Value)
			if err != nil {
				t.Errorf(
					"FindValidByValue(ctx, %v) (*auth.RefreshToken, error) = %v, %v",
					otherToken.Value, found, err,
				)
			}
		},
	)

	t.Run(
		"returns no error when the user has no refresh tokens",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx := context.Background()

			userID := user.ID(1)

			if err := rtr.RevokeAllForUser(ctx, userID); err != nil {
				t.Fatalf(
					"RevokeAllForUser(ctx, %d) error = %v",
					userID, err,
				)
			}
		},
	)

	t.Run(
		"returns a wrapped error when the context is canceled",
		func(t *testing.T) {
			rtr := newTestRepo(t, time.Hour)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			userID := user.ID(1)

			err := rtr.RevokeAllForUser(ctx, userID)
			if !errors.Is(err, context.Canceled) {
				t.Fatalf(
					"RevokeAllForUser(ctx, %d) error = %v, want %v",
					userID, err, context.Canceled,
				)
			}
		},
	)
}
