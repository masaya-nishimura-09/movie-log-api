package record

import "testing"

func TestNewRuntime(t *testing.T) {
	tests := []struct {
		name    string
		input   uint
		want    Runtime
		wantErr bool
	}{
		{"zero", 0, 0, false},
		{"max", 1440, 1440, false},

		{"over max", 1441, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRuntime(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewRuntime(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewRuntime(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
