package movie

import "testing"

func TestNewPosterURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    PosterURL
		wantErr bool
	}{
		{
			"https url",
			"https://image.tmdb.org/t/p/w500/6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg",
			"https://image.tmdb.org/t/p/w500/6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg",
			false,
		},
		{"http url", "http://example.com/poster.jpg", "http://example.com/poster.jpg", false},
		{"empty", "", "", false},

		{"tmdb relative path", "/6FfCtAuVAW8XJjZ7eWeLibRLWTw.jpg", "", true},
		{"other scheme", "ftp://example.com/poster.jpg", "", true},
		{"no host", "https://", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPosterURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPosterURL(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewPosterURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
