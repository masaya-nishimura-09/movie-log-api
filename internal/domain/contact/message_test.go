package contact

import (
	"strings"
	"testing"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Message
		wantErr bool
	}{
		{"typical", "退会できません", "退会できません", false},
		{"trimmed", "  退会できません\n", "退会できません", false},
		{"keeps inner newlines", "1行目\n2行目", "1行目\n2行目", false},
		{"max length", strings.Repeat("a", 2000), Message(strings.Repeat("a", 2000)), false},
		{"japanese at max length", strings.Repeat("あ", 2000), Message(strings.Repeat("あ", 2000)), false},

		{"empty", "", "", true},
		{"only spaces", " \n\t", "", true},
		{"over max length", strings.Repeat("a", 2001), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMessage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewMessage(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewMessage(len=%d) = %q, want %q", len(tt.input), got, tt.want)
			}
		})
	}
}
