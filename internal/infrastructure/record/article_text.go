package record

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
)

func extractPlainText(content record.ArticleContent) (string, error) {
	var root record.ArticleContentNode
	if err := json.Unmarshal(content, &root); err != nil {
		return "", fmt.Errorf("unmarshal article content: %w", err)
	}

	return strings.Join(articleTextLines(root), "\n"), nil
}

func articleTextLines(n record.ArticleContentNode) []string {
	if articleHasInlineText(n) {
		return []string{articleInlineText(n)}
	}

	var lines []string
	for _, child := range n.Content {
		lines = append(lines, articleTextLines(child)...)
	}
	return lines
}

func articleHasInlineText(n record.ArticleContentNode) bool {
	for _, child := range n.Content {
		if child.Type == "text" {
			return true
		}
	}
	return false
}

func articleInlineText(n record.ArticleContentNode) string {
	var sb strings.Builder
	for _, child := range n.Content {
		switch child.Type {
		case "text":
			sb.WriteString(child.Text)
		case "hardBreak":
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
