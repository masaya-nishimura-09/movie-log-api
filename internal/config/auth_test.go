package config

import (
	"strings"
	"testing"
)

func TestSecret(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "returns an error when JWT_SECRET is empty", value: "", wantErr: true},
		{name: "returns an error when JWT_SECRET is shorter than 32 bytes", value: strings.Repeat("a", 31), wantErr: true},
		{name: "returns the secret when JWT_SECRET is 32 bytes", value: strings.Repeat("a", 32), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("JWT_SECRET", tt.value)

			got, err := Secret()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Secret() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && string(got) != tt.value {
				t.Errorf("Secret() = %q, want %q", got, tt.value)
			}
		})
	}
}
