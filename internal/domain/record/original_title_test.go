package record

import (
	"strings"
	"testing"
)

func TestNewOriginalTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    OriginalTitle
		wantErr bool
	}{
		{"typical", "Arrival", "Arrival", false},
		{"empty", "", "", false},
		{"max length", strings.Repeat("a", 255), OriginalTitle(strings.Repeat("a", 255)), false},
		{"japanese at max length", strings.Repeat("あ", 255), OriginalTitle(strings.Repeat("あ", 255)), false},

		{"over max length", strings.Repeat("a", 256), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOriginalTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewOriginalTitle(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewOriginalTitle(len=%d) = %q, want %q", len(tt.input), got, tt.want)
			}
		})
	}
}
