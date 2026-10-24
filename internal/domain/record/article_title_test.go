package record

import (
	"strings"
	"testing"
)

func TestNewArticleTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    ArticleTitle
		wantErr bool
	}{
		{"typical", "Pulp Fiction を観て", "Pulp Fiction を観て", false},
		{"max length", strings.Repeat("a", 255), ArticleTitle(strings.Repeat("a", 255)), false},
		{"japanese at max length", strings.Repeat("あ", 255), ArticleTitle(strings.Repeat("あ", 255)), false},

		{"empty", "", "", true},
		{"over max length", strings.Repeat("a", 256), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewArticleTitle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewArticleTitle(len=%d) error = %v, wantErr %v", len(tt.input), err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("NewArticleTitle(len=%d) = %q, want %q", len(tt.input), got, tt.want)
			}
		})
	}
}
