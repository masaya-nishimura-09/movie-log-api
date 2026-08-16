package record

import (
	"strings"
	"testing"
)

func TestNewMemo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Memo
		wantErr bool
	}{
		{"typical", "面白かった", "面白かった", false},
		{"empty", "", "", false},
		{"max length", strings.Repeat("a", 1000), Memo(strings.Repeat("a", 1000)), false},
		{"japanese at max length", strings.Repeat("あ", 1000), Memo(strings.Repeat("あ", 1000)), false},

		{"over max length", strings.Repeat("a", 1001), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMemo(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewMemo(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewMemo(len=%d) = %q, want %q", len(tt.input), got, tt.want)
			}
		})
	}
}
