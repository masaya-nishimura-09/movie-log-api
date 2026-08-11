package user

import (
	"strings"
	"testing"
)

func TestNewPassword(t *testing.T) {
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
			_, err := NewPassword(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPassword(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
