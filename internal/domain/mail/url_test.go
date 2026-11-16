package mail

import "testing"

func TestNewURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    URL
		wantErr bool
	}{
		{"https url", "https://www.cinelier.app/ja/login", "https://www.cinelier.app/ja/login", false},
		{"http url", "http://localhost:3000/ja/login", "http://localhost:3000/ja/login", false},

		{"empty", "", "", true},
		{"relative path", "/ja/login", "", true},
		{"other scheme", "ftp://example.com/", "", true},
		{"no host", "https://", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewURL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
