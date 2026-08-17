package record

import "testing"

func TestNewLanguage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Language
		wantErr bool
	}{
		{"iso 639-1 code", "ja", "ja", false},
		{"uppercase", "EN", "en", false},
		{"iso 639-2 code", "jpn", "ja", false},
		{"undetermined", "und", "und", false},

		{"empty", "", "", true},
		{"unknown code", "zz", "", true},
		{"language name", "english", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLanguage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewLanguage(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewLanguage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
