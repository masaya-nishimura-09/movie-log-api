package media

import "testing"

func TestNewData(t *testing.T) {
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
			got, err := NewData(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewData(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != len(tt.input) {
				t.Errorf("NewData(len=%d) length = %d, want %d", len(tt.input), len(got), len(tt.input))
			}
			if !tt.wantErr {
				tt.input[0] = 0x00
				if got[0] == 0x00 {
					t.Errorf("NewData(len=%d) data is not copied", len(tt.input))
				}
			}
		})
	}
}

func TestNewContentType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ContentType
		wantErr bool
	}{
		{"jpeg", "image/jpeg", ContentTypeJPEG, false},
		{"png", "image/png", ContentTypePNG, false},
		{"webp", "image/webp", ContentTypeWebP, false},

		{"empty", "", "", true},
		{"undefined value", "image/gif", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewContentType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewContentType(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewContentType(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewMedia(t *testing.T) {
	tests := []struct {
		name    string
		input   Data
		want    ContentType
		wantErr bool
	}{
		{"jpeg", Data("\xFF\xD8\xFF"), ContentTypeJPEG, false},
		{"png", Data("\x89PNG\x0D\x0A\x1A\x0A"), ContentTypePNG, false},
		{"webp", Data("RIFF____WEBPVP"), ContentTypeWebP, false},

		{"undefined content type", Data("GIF89a"), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMedia(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewMedia(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got.ContentType != tt.want {
				t.Errorf("NewMedia(len=%d) content type = %q, want %q", len(tt.input), got.ContentType, tt.want)
			}
		})
	}
}
