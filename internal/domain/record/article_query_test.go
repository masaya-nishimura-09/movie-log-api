package record

import (
	"strings"
	"testing"
)

func TestNewArticleKeyword(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ArticleKeyword
		wantErr bool
	}{
		{"empty", "", "", false},
		{"typical", "感想", "感想", false},
		{"max length", strings.Repeat("a", 255), ArticleKeyword(strings.Repeat("a", 255)), false},
		{"japanese at max length", strings.Repeat("あ", 255), ArticleKeyword(strings.Repeat("あ", 255)), false},

		{"over max length", strings.Repeat("a", 256), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewArticleKeyword(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewArticleKeyword(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewArticleKeyword(len=%d) = %q, want %q", len(tt.input), got, tt.want)
			}
		})
	}
}

func TestNewArticleSortField(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ArticleSortField
		wantErr bool
	}{
		{"created_at", "created_at", ArticleSortFieldCreatedAt, false},
		{"updated_at", "updated_at", ArticleSortFieldUpdatedAt, false},
		{"title", "title", ArticleSortFieldTitle, false},

		{"empty", "", "", true},
		{"invalid", "invalid", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewArticleSortField(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewArticleSortField(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewArticleSortField(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
