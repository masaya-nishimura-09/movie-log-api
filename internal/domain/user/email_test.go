package user

import "testing"

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Email
		wantErr bool
	}{
		{"valid", "test@example.com", "test@example.com", false},
		{"valid with subdomain", "test@mail.example.com", "test@mail.example.com", false},
		{"valid with plus", "test+tag@example.com", "test+tag@example.com", false},
		{"valid with dot in local part", "first.last@example.com", "first.last@example.com", false},
		{"normalized to lowercase", "Test@Example.COM", "test@example.com", false},

		{"empty", "", "", true},
		{"no at sign", "testexample.com", "", true},
		{"no domain", "test@", "", true},
		{"no local part", "@example.com", "", true},
		{"no tld", "test@example", "", true},
		{"double at sign", "test@@example.com", "", true},
		{"space in email", "test @example.com", "", true},
		{"trailing dot in domain", "test@example.com.", "", true},
		{"leading dot in domain", "test@.example.com", "", true},
		{"consecutive dots in domain", "test@example..com", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmail(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewEmail(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewEmail(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
