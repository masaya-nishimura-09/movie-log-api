package user

import (
	"strings"
	"testing"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Username
		wantErr  bool
	}{
		{"valid", "Test", Username("Test"), false},
		{"empty", "", Username(""), true},
		{"too long", strings.Repeat("a", 101), Username(""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			username, err := NewUsername(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					`NewUsername(%q) (username, error) = %s, %v, wantErr %v`,
					tt.input,
					err,
					tt.wantErr,
				)
			}
			if username != tt.expected {
				t.Errorf(
					`NewUsername(%q) (username, error) = %s, %v, wantErr %v`,
					tt.input,
					err,
					tt.wantErr,
				)
			}
		})
	}
}
