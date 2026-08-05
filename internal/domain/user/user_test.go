package user

import (
	"strings"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "Test", false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 101), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(Username(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf(`ValidateUsername(%q) error = %v, wantErr %v`, tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid", "test@example.com", false},
		{"valid with subdomain", "test@mail.example.com", false},
		{"valid with plus", "test+tag@example.com", false},
		{"valid with dot in local part", "first.last@example.com", false},

		{"empty", "", true},
		{"no at sign", "testexample.com", true},
		{"no domain", "test@", true},
		{"no local part", "@example.com", true},
		{"no tld", "test@example", true},
		{"double at sign", "test@@example.com", true},
		{"space in email", "test @example.com", true},
		{"trailing dot in domain", "test@example.com.", true},
		{"leading dot in domain", "test@.example.com", true},
		{"consecutive dots in domain", "test@example..com", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(Email(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf(`ValidateEmail(%q) error = %v, wantErr %v`, tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid minimum length", strings.Repeat("a", 8), false},
		{"valid maximum length", strings.Repeat("a", 72), false},
		{"empty", "", true},
		{"too short", strings.Repeat("a", 7), true},
		{"too long", strings.Repeat("a", 73), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(Password(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf(`ValidatePassword(%q) error = %v, wantErr %v`, tt.input, err, tt.wantErr)
			}
		})
	}
}
