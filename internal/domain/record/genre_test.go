package record

import (
	"slices"
	"testing"
)

func TestNewGenre(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Genre
		wantErr bool
	}{
		{"defined value", "action", GenreAction, false},

		{"empty", "", "", true},
		{"undefined value", "musical", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGenre(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewGenre(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewGenre(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewGenres(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []Genre
		wantErr bool
	}{
		{"multiple values", []string{"action", "crime"}, []Genre{GenreAction, GenreCrime}, false},
		{"empty slice", []string{}, []Genre{}, false},

		{"duplicate value", []string{"action", "action"}, nil, true},
		{"undefined element", []string{"action", "musical"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGenres(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewGenres(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("NewGenres(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
