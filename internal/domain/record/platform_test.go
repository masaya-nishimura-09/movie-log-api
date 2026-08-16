package record

import "testing"

func TestNewPlatform(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Platform
		wantErr bool
	}{
		{"defined value", "netflix", PlatformNetflix, false},
		{"other", "other", PlatformOther, false},

		{"empty", "", "", true},
		{"undefined value", "disney", "", true},
		{"different case", "Netflix", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPlatform(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPlatform(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewPlatform(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
