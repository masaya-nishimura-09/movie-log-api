package mail

import "testing"

func TestNewReplyTo(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ReplyTo
		wantErr bool
	}{
		{"valid", "test@example.com", "test@example.com", false},
		{"normalized to lowercase", "Test@Example.COM", "test@example.com", false},
		{"empty", "", "", false},

		{"no at sign", "testexample.com", "", true},
		{"no domain", "test@", "", true},
		{"no tld", "test@example", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewReplyTo(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewReplyTo(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewReplyTo(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
