package movie

import "testing"

func TestNewLanguage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DisplayLanguage
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
			got, err := NewDisplayLanguage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewDisplayLanguage(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewDisplayLanguage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
