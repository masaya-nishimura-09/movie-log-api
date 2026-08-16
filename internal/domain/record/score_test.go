package record

import "testing"

func TestNewScore(t *testing.T) {
	tests := []struct {
		name    string
		input   uint
		want    Score
		wantErr bool
	}{
		{"min", 1, 1, false},
		{"max", 5, 5, false},

		{"zero", 0, 0, true},
		{"over max", 6, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewScore(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewScore(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewScore(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
