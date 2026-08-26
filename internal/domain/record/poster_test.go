package record

import "testing"

func TestNewPosterData(t *testing.T) {
	jpeg := []byte("\xFF\xD8\xFF")

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{"typical", jpeg, false},
		{"max size", append(jpeg, make([]byte, 5*1024*1024-len(jpeg))...), false},

		{"empty", nil, true},
		{"over max size", append(jpeg, make([]byte, 5*1024*1024-len(jpeg)+1)...), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPosterData(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPosterData(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != len(tt.input) {
				t.Errorf("NewPosterData(len=%d) length = %d, want %d", len(tt.input), len(got), len(tt.input))
			}
			if !tt.wantErr {
				tt.input[0] = 0x00
				if got[0] == 0x00 {
					t.Errorf("NewPosterData(len=%d) data is not copied", len(tt.input))
				}
			}
		})
	}
}

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
	tests := []struct {
		name    string
		input   PosterData
		want    PosterContentType
		wantErr bool
	}{
		{"jpeg", PosterData("\xFF\xD8\xFF"), PosterContentTypeJPEG, false},
		{"png", PosterData("\x89PNG\x0D\x0A\x1A\x0A"), PosterContentTypePNG, false},
		{"webp", PosterData("RIFF____WEBPVP"), PosterContentTypeWebP, false},

		{"undefined content type", PosterData("GIF89a"), "", true},
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
		})
	}
}
