package user

import "testing"

func TestNewEmailVerificationTokenValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    EmailVerificationTokenValue
		wantErr bool
	}{
		{"typical", "abc123", "abc123", false},

		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmailVerificationTokenValue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewEmailVerificationTokenValue(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewEmailVerificationTokenValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
