package movie

import (
	"testing"
	"time"
)

func TestNewReleaseYear(t *testing.T) {
	currentYear := uint(time.Now().Year())

	tests := []struct {
		name    string
		input   uint
		want    ReleaseYear
		wantErr bool
	}{
		{"min", 1888, 1888, false},
		{"max", currentYear + 5, ReleaseYear(currentYear + 5), false},

		{"before min", 1887, 0, true},
		{"over max", currentYear + 6, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewReleaseYear(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewReleaseYear(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewReleaseYear(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
