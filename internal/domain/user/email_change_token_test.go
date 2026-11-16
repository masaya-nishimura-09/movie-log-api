package user

import "testing"

func TestNewEmailChangeTokenValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    EmailChangeTokenValue
		wantErr bool
	}{
		{"typical", "abc123", "abc123", false},

		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmailChangeTokenValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewEmailChangeTokenValue(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewEmailChangeTokenValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
