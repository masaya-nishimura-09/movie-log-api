package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/auth"
	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/user"
)

func TestAccessTokenService_GenerateAndValidate(t *testing.T) {
	svc := NewAccessTokenService([]byte("test-secret"), time.Hour)
	principal := auth.Principal{UserID: user.ID(42), Role: user.RoleAdmin}

	token, err := svc.Generate(context.Background(), &principal)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if token.Value == "" {
		t.Fatal("Generate() returned an empty token value")
	}

	got, err := svc.Validate(context.Background(), token)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got.UserID != principal.UserID || got.Role != principal.Role {
		t.Errorf("Validate() = %+v, want %+v", got, principal)
	}
}

func TestAccessTokenService_Validate_ExpiredToken(t *testing.T) {
	svc := NewAccessTokenService([]byte("test-secret"), -time.Hour) // already expired
	principal := auth.Principal{UserID: user.ID(1), Role: user.RoleUser}

	token, err := svc.Generate(context.Background(), &principal)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := svc.Validate(context.Background(), token); err == nil {
		t.Fatal("Validate() error = nil, want an error for an expired token")
	}
}

func TestAccessTokenService_Validate_TamperedToken(t *testing.T) {
	svc := NewAccessTokenService([]byte("test-secret"), time.Hour)
	principal := auth.Principal{UserID: user.ID(1), Role: user.RoleUser}

	token, err := svc.Generate(context.Background(), &principal)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	tampered := &auth.AccessToken{
		Value: auth.AccessTokenValue(string(token.Value) + "tampered"),
	}
	if _, err := svc.Validate(context.Background(), tampered); err == nil {
		t.Fatal("Validate() error = nil, want an error for a tampered token")
	}
}

func TestAccessTokenService_Validate_RejectsUnexpectedAlgorithm(t *testing.T) {
	secret := []byte("test-secret")
	svc := NewAccessTokenService(secret, time.Hour)

	c := claims{
		UserID: user.ID(1),
		Role:   user.RoleUser,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	// Sign with the same secret but a different algorithm than the one the
	// service issues (HS256), to prove Validate rejects it regardless of the
	// signature being otherwise valid.
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS384, c).SignedString(secret)
	if err != nil {
		t.Fatalf("failed to build test token: %v", err)
	}

	token := &auth.AccessToken{Value: auth.AccessTokenValue(tokenStr)}
	if _, err := svc.Validate(context.Background(), token); err == nil {
		t.Fatal("Validate() error = nil, want an error for a token signed with an unexpected algorithm")
	}
}

func TestAccessTokenService_Validate_MalformedToken(t *testing.T) {
	svc := NewAccessTokenService([]byte("test-secret"), time.Hour)
	token := &auth.AccessToken{Value: "not-a-jwt"}

	if _, err := svc.Validate(context.Background(), token); err == nil {
		t.Fatal("Validate() error = nil, want an error for a malformed token")
	}
}
