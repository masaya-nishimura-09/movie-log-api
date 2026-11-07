package record

import (
	"strings"
	"testing"
)

func TestNewOverview(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Overview
		wantErr bool
	}{
		{"typical", "言語学者が異星人と対話する", "言語学者が異星人と対話する", false},
		{"empty", "", "", false},
		{"max length", strings.Repeat("a", 2000), Overview(strings.Repeat("a", 2000)), false},
		{"japanese at max length", strings.Repeat("あ", 2000), Overview(strings.Repeat("あ", 2000)), false},

		{"over max length", strings.Repeat("a", 2001), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewOverview(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewOverview(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewOverview(len=%d) = %q, want %q", len(tt.input), got, tt.want)
			}
		})
	}
}
