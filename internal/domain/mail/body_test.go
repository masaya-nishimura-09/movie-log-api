package mail

import "testing"

func TestNewHTMLBody(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    HTMLBody
		wantErr bool
	}{
		{"typical", "<p>本文</p>", "<p>本文</p>", false},

		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewHTMLBody(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewHTMLBody(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewHTMLBody(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewTextBody(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TextBody
		wantErr bool
	}{
		{"typical", "本文", "本文", false},

		{"empty", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTextBody(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewTextBody(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewTextBody(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
