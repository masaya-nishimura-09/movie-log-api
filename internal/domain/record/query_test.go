package record

import (
	"strings"
	"testing"
)

func TestNewTitleKeyword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TitleKeyword
		wantErr bool
	}{
		{"empty", "", "", false},
		{"typical", "Pulp Fiction", "Pulp Fiction", false},
		{"max length", strings.Repeat("a", 255), TitleKeyword(strings.Repeat("a", 255)), false},
		{"japanese at max length", strings.Repeat("あ", 255), TitleKeyword(strings.Repeat("あ", 255)), false},

		{"over max length", strings.Repeat("a", 256), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTitleKeyword(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewTitleKeyword(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewTitleKeyword(len=%d) = %q, want %q", len(tt.input), got, tt.want)
			}
		})
	}
}

func TestNewSortField(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    SortField
		wantErr bool
	}{
		{"watched_at", "watched_at", SortFieldWatchedAt, false},
		{"release_year", "release_year", SortFieldReleaseYear, false},
		{"score", "score", SortFieldScore, false},
		{"title", "title", SortFieldTitle, false},

		{"empty", "", "", true},
		{"invalid", "invalid", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSortField(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewSortField(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewSortField(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewSortOrder(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    SortOrder
		wantErr bool
	}{
		{"asc", "asc", SortOrderAsc, false},
		{"desc", "desc", SortOrderDesc, false},

		{"empty", "", "", true},
		{"invalid", "invalid", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSortOrder(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewSortOrder(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewSortOrder(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewPage(t *testing.T) {
	tests := []struct {
		name    string
		input   uint
		want    Page
		wantErr bool
	}{
		{"min", 1, 1, false},
		{"typical", 10, 10, false},

		{"zero", 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPage(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewPage(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewPerPage(t *testing.T) {
	tests := []struct {
		name    string
		input   uint
		want    PerPage
		wantErr bool
	}{
		{"min", 1, 1, false},
		{"typical", 20, 20, false},
		{"max", 100, 100, false},

		{"zero", 0, 0, true},
		{"over max", 101, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPerPage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewPerPage(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewPerPage(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
