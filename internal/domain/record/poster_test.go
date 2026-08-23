package record

import "testing"

func TestNewPosterContentType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    PosterContentType
		wantErr bool
	}{
		{"jpeg", "image/jpeg", PosterContentTypeJPEG, false},
		{"png", "image/png", PosterContentTypePNG, false},
		{"webp", "image/webp", PosterContentTypeWebP, false},

		{"empty", "", "", true},
		{"undefined value", "image/gif", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPosterContentType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPosterContentType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewPosterContentType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewPoster(t *testing.T) {
	jpeg := []byte("\xFF\xD8\xFF")
	png := []byte("\x89PNG\x0D\x0A\x1A\x0A")
	webp := []byte("RIFF____WEBPVP")
	gif := []byte("GIF89a")

	tests := []struct {
		name    string
		input   []byte
		want    PosterContentType
		wantErr bool
	}{
		{"jpeg", jpeg, PosterContentTypeJPEG, false},
		{"png", png, PosterContentTypePNG, false},
		{"webp", webp, PosterContentTypeWebP, false},
		{"max size", append(jpeg, make([]byte, 5*1024*1024-len(jpeg))...), PosterContentTypeJPEG, false},

		{"empty", nil, "", true},
		{"over max size", append(jpeg, make([]byte, 5*1024*1024-len(jpeg)+1)...), "", true},
		{"undefined content type", gif, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPoster(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPoster(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got.ContentType != tt.want {
				t.Errorf("NewPoster(len=%d) content type = %q, want %q", len(tt.input), got.ContentType, tt.want)
			}
			if !tt.wantErr && len(got.Data) != len(tt.input) {
				t.Errorf("NewPoster(len=%d) data length = %d, want %d", len(tt.input), len(got.Data), len(tt.input))
			}
			if !tt.wantErr {
				tt.input[0] = 0x00
				if got.Data[0] == 0x00 {
					t.Errorf("NewPoster(len=%d) data is not copied", len(tt.input))
				}
			}
		})
	}
}
