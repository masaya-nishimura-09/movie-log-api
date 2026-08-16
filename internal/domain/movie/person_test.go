package movie

import (
	"strings"
	"testing"
)

func TestNewPersonName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    PersonName
		wantErr bool
	}{
		{"typical", "Quentin Tarantino", "Quentin Tarantino", false},
		{"max length", strings.Repeat("a", 100), PersonName(strings.Repeat("a", 100)), false},
		{"japanese at max length", strings.Repeat("あ", 100), PersonName(strings.Repeat("あ", 100)), false},

		{"empty", "", "", true},
		{"over max length", strings.Repeat("a", 101), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPersonName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPersonName(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewPersonName(len=%d) = %q, want %q", len(tt.input), got, tt.want)
			}
		})
	}
}
