package record

import (
	"testing"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
)

func TestExtractPlainText(t *testing.T) {
	tests := []struct {
		name    string
		input   record.ArticleContent
		want    string
		wantErr bool
	}{
		{
			"single paragraph",
			record.ArticleContent(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"感想です"}]}]}`),
			"感想です",
			false,
		},
		{
			"paragraph split by marks",
			record.ArticleContent(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"この映画の"},{"type":"text","marks":[{"type":"bold"}],"text":"感想"},{"type":"text","text":"を書く"}]}]}`),
			"この映画の感想を書く",
			false,
		},
		{
			"multiple paragraphs",
			record.ArticleContent(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"一つ目"}]},{"type":"paragraph","content":[{"type":"text","text":"二つ目"}]}]}`),
			"一つ目\n二つ目",
			false,
		},
		{
			"heading and nested list",
			record.ArticleContent(`{"type":"doc","content":[{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"見出し"}]},{"type":"bulletList","content":[{"type":"listItem","content":[{"type":"paragraph","content":[{"type":"text","text":"項目"}]}]}]}]}`),
			"見出し\n項目",
			false,
		},
		{
			"hard break",
			record.ArticleContent(`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"りんご"},{"type":"hardBreak"},{"type":"text","text":"みかん"}]}]}`),
			"りんご\nみかん",
			false,
		},
		{
			"empty paragraph is skipped",
			record.ArticleContent(`{"type":"doc","content":[{"type":"paragraph"},{"type":"paragraph","content":[{"type":"text","text":"本文"}]}]}`),
			"本文",
			false,
		},
		{"doc without content", record.ArticleContent(`{"type":"doc"}`), "", false},

		{"invalid json", record.ArticleContent(`{"type":`), "", true},
		{"array", record.ArticleContent(`[1,2]`), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractPlainText(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("extractPlainText(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("extractPlainText(%s) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
