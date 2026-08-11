package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

func TestGenerate(t *testing.T) {
	t.Run(
		"returns an access token carrying the principal and expiring after the ttl",
		func(t *testing.T) {
			secret := []byte("testsecret")
			ttl := time.Hour
			ats := NewAccessTokenService(secret, ttl)

			ctx := context.Background()
			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			issuedBefore := time.Now()
			at, err := ats.Generate(ctx, &principal)
			if err != nil {
				t.Fatalf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, %v",
					principal, at, err,
				)
			}
			issuedAfter := time.Now()

			if at.Value == "" {
				t.Errorf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, want non-empty Value",
					principal, at,
				)
			}
			if at.Principal != principal {
				t.Errorf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, want Principal %v",
					principal, at, principal,
				)
			}

			wantMin := issuedBefore.Add(ttl)
			wantMax := issuedAfter.Add(ttl)
			if at.ExpiresAt.Before(wantMin) || at.ExpiresAt.After(wantMax) {
				t.Errorf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, want ExpiresAt in [%v, %v]",
					principal, at, wantMin, wantMax,
				)
			}

			parsed, err := jwt.ParseWithClaims(
				string(at.Value),
				&claims{},
				func(t *jwt.Token) (any, error) { return secret, nil },
			)
			if err != nil {
				t.Fatalf(
					"ParseWithClaims(%v) (*jwt.Token, error) = %v, %v",
					at.Value, parsed, err,
				)
			}

			if parsed.Method.Alg() != jwt.SigningMethodHS256.Name {
				t.Errorf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, want alg %v, got %v",
					principal, at, jwt.SigningMethodHS256.Name, parsed.Method.Alg(),
				)
			}

			got := parsed.Claims.(*claims)
			if got.UserID != principal.UserID || got.Role != principal.Role {
				t.Errorf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, want claims %v, got %v",
					principal, at, principal, got,
				)
			}

			if got.IssuedAt == nil ||
				got.IssuedAt.Before(issuedBefore.Truncate(time.Second)) ||
				got.IssuedAt.After(issuedAfter) {
				t.Errorf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, want IssuedAt in [%v, %v], got %v",
					principal, at, issuedBefore, issuedAfter, got.IssuedAt,
				)
			}

			if got.ExpiresAt == nil {
				t.Fatalf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, want non-nil exp",
					principal, at,
				)
			}
			if !at.ExpiresAt.Truncate(time.Second).Equal(got.ExpiresAt.Time) {
				t.Errorf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, want exp %v, got %v",
					principal, at, at.ExpiresAt.Truncate(time.Second), got.ExpiresAt.Time,
				)
			}
		},
	)
}

func TestValidate(t *testing.T) {
	t.Run(
		"returns the principal when a token from Generate is given",
		func(t *testing.T) {
			secret := []byte("testsecret")
			ats := NewAccessTokenService(secret, time.Hour)

			ctx := context.Background()
			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			at, err := ats.Generate(ctx, &principal)
			if err != nil {
				t.Fatalf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, %v",
					principal, at, err,
				)
			}

			got, err := ats.Validate(ctx, at)
			if err != nil {
				t.Fatalf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, %v",
					at, got, err,
				)
			}
			if got.UserID != principal.UserID || got.Role != principal.Role {
				t.Errorf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, want %v",
					at, got, principal,
				)
			}
		},
	)

	t.Run(
		"returns ErrTokenSignatureInvalid when the token is signed with a different secret",
		func(t *testing.T) {
			issuer := NewAccessTokenService([]byte("othersecret"), time.Hour)
			ats := NewAccessTokenService([]byte("testsecret"), time.Hour)

			ctx := context.Background()
			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			at, err := issuer.Generate(ctx, &principal)
			if err != nil {
				t.Fatalf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, %v",
					principal, at, err,
				)
			}

			got, err := ats.Validate(ctx, at)
			if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				t.Fatalf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, %v, want %v",
					at, got, err, jwt.ErrTokenSignatureInvalid,
				)
			}
			if got != nil {
				t.Errorf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, %v, want nil",
					at, got, err,
				)
			}
		},
	)

	t.Run(
		"returns ErrTokenExpired when the token is expired",
		func(t *testing.T) {
			secret := []byte("testsecret")
			issuer := NewAccessTokenService(secret, -time.Hour)
			ats := NewAccessTokenService(secret, time.Hour)

			ctx := context.Background()
			principal := auth.Principal{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
			}

			at, err := issuer.Generate(ctx, &principal)
			if err != nil {
				t.Fatalf(
					"Generate(ctx, %v) (*auth.AccessToken, error) = %v, %v",
					principal, at, err,
				)
			}

			got, err := ats.Validate(ctx, at)
			if !errors.Is(err, jwt.ErrTokenExpired) {
				t.Fatalf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, %v, want %v",
					at, got, err, jwt.ErrTokenExpired,
				)
			}
			if got != nil {
				t.Errorf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, %v, want nil",
					at, got, err,
				)
			}
		},
	)

	t.Run(
		"returns ErrTokenSignatureInvalid when the token is signed with HS512",
		func(t *testing.T) {
			secret := []byte("testsecret")
			ats := NewAccessTokenService(secret, time.Hour)

			ctx := context.Background()
			c := claims{
				UserID: user.ID(1),
				Role:   user.RoleAdmin,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			}
			tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS512, c).SignedString(secret)
			if err != nil {
				t.Fatalf(
					"SignedString(secret) (string, error) = %v, %v",
					tokenStr, err,
				)
			}
			at := auth.AccessToken{Value: auth.AccessTokenValue(tokenStr)}

			got, err := ats.Validate(ctx, &at)
			if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				t.Fatalf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, %v, want %v",
					at, got, err, jwt.ErrTokenSignatureInvalid,
				)
			}
			if got != nil {
				t.Errorf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, %v, want nil",
					at, got, err,
				)
			}
		},
	)

	t.Run(
		"returns ErrTokenMalformed when the token string is malformed",
		func(t *testing.T) {
			ats := NewAccessTokenService([]byte("testsecret"), time.Hour)

			ctx := context.Background()
			at := auth.AccessToken{Value: auth.AccessTokenValue("not.a.token")}

			got, err := ats.Validate(ctx, &at)
			if !errors.Is(err, jwt.ErrTokenMalformed) {
				t.Fatalf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, %v, want %v",
					at, got, err, jwt.ErrTokenMalformed,
				)
			}
			if got != nil {
				t.Errorf(
					"Validate(ctx, %v) (*auth.Principal, error) = %v, %v, want nil",
					at, got, err,
				)
			}
		},
	)
}
