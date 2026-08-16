package user

import (
	"strings"
	"testing"
)

func TestNewUsername(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Username
		wantErr bool
	}{
		{"valid", "Test", "Test", false},
		{"multibyte at limit", strings.Repeat("あ", 100), Username(strings.Repeat("あ", 100)), false},
		{"empty", "", "", true},
		{"too long", strings.Repeat("a", 101), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewUsername(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewUsername(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewUsername(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
