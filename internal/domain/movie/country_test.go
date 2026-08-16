package movie

import (
	"slices"
	"testing"
)

func TestNewCountry(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Country
		wantErr bool
	}{
		{"alpha-2 code", "JP", "JP", false},
		{"lowercase", "jp", "JP", false},
		{"alpha-3 code", "JPN", "JP", false},

		{"empty", "", "", true},
		{"unassigned code", "ZZ", "", true},
		{"region grouping", "EU", "", true},
		{"country name", "Japan", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCountry(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCountry(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewCountry(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewCountries(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []Country
		wantErr bool
	}{
		{"multiple values", []string{"JP", "US"}, []Country{"JP", "US"}, false},
		{"empty slice", []string{}, []Country{}, false},

		{"duplicate value", []string{"JP", "JP"}, nil, true},
		{"duplicate after normalization", []string{"JP", "jpn"}, nil, true},
		{"invalid element", []string{"JP", "ZZ"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewCountries(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewCountries(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("NewCountries(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
