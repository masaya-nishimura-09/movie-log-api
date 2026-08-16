package record

import "testing"

func TestNewWatchCount(t *testing.T) {
	tests := []struct {
		name    string
		input   uint
		want    WatchCount
		wantErr bool
	}{
		{"min", 1, 1, false},
		{"repeated viewing", 5, 5, false},

		{"zero", 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWatchCount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewWatchCount(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewWatchCount(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
