package record

import (
	"strings"
	"testing"
)

func TestNewTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Title
		wantErr bool
	}{
		{"typical", "Pulp Fiction", "Pulp Fiction", false},
		{"max length", strings.Repeat("a", 255), Title(strings.Repeat("a", 255)), false},
		{"japanese at max length", strings.Repeat("あ", 255), Title(strings.Repeat("あ", 255)), false},

		{"empty", "", "", true},
		{"over max length", strings.Repeat("a", 256), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewTitle(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewTitle(len=%d) = %q, want %q", len(tt.input), got, tt.want)
			}
		})
	}
}
