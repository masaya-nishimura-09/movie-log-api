package mail

import "testing"

func TestNewLanguage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Language
		wantErr bool
	}{
		{"japanese", "ja", LanguageJA, false},
		{"english", "en", LanguageEN, false},

		{"empty", "", "", true},
		{"unsupported", "fr", "", true},
		{"uppercase", "JA", "", true},
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
