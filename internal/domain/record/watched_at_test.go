package record

import (
	"testing"
	"time"
)

func TestNewWatchedAt(t *testing.T) {
	yesterday := time.Now().Add(-24 * time.Hour)
	longPast := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	tomorrow := time.Now().Add(24 * time.Hour)

	tests := []struct {
		name    string
		input   time.Time
		want    time.Time
		wantErr bool
	}{
		{"yesterday", yesterday, yesterday, false},
		{"long past", longPast, longPast, false},

		{"zero value", time.Time{}, time.Time{}, true},
		{"future", tomorrow, time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWatchedAt(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewWatchedAt(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !got.Equal(tt.want) {
				t.Errorf("NewWatchedAt(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
