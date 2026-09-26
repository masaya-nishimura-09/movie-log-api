package record

import "testing"

func TestNewTotalCount(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		want    TotalCount
		wantErr bool
	}{
		{"zero", 0, 0, false},
		{"positive", 100, 100, false},

		{"negative", -1, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTotalCount(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewTotalCount(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewTotalCount(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
