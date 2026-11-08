package record

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/record"
)

func extractPlainText(note record.Note) (string, error) {
	var root record.NoteNode
	if err := json.Unmarshal(note, &root); err != nil {
		return "", fmt.Errorf("unmarshal note: %w", err)
	}

	return strings.Join(noteTextLines(root), "\n"), nil
}

func noteTextLines(n record.NoteNode) []string {
	if noteHasInlineText(n) {
		return []string{noteInlineText(n)}
	}

	var lines []string
	for _, child := range n.Content {
		lines = append(lines, noteTextLines(child)...)
	}
	return lines
}

func noteHasInlineText(n record.NoteNode) bool {
	for _, child := range n.Content {
		if child.Type == "text" {
			return true
		}
	}
	return false
}

func noteInlineText(n record.NoteNode) string {
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
