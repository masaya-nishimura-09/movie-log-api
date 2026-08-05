package repository

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestMain(m *testing.M) {
	if err := Init(); err != nil {
		panic(err)
	}
	m.Run()
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name    string
		userID  uint
		role    string
		wantErr bool
	}{
		{"valid", 1, "user", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := Generate(tt.userID, tt.role)
			if (token == "" || err != nil) != tt.wantErr {
				t.Errorf(
					`Generate(%d, %q) (string, error) = (%q, %v)`,
					tt.userID,
					tt.role,
					token,
					err,
				)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	validToken, err := Generate(1, "user")
	if err != nil {
		t.Fatalf("failed to create valid token: %v", err)
	}

	expiredC := Claims{
		UserID: 1,
		Role:   "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now()),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	expiredT := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredC)
	expiredTokenStr, err := expiredT.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	tests := []struct {
		name     string
		tokenStr string
		wantErr  bool
	}{
		{"valid", validToken, false},
		{"empty", "", true},
		{"invalid token", "12345", true},
		{"expired", expiredTokenStr, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := Validate(tt.tokenStr)
			if (claims == nil || err != nil) != tt.wantErr {
				t.Errorf(
					`Validate(%q) (*Claims, error) = (%v, %v)`,
					tt.tokenStr,
					claims,
					err,
				)
			}
		})
	}
}
