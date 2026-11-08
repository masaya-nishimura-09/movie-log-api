package record

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Note json.RawMessage

const NoteMaxBytes = 100 * 1024

func NewNote(value json.RawMessage) (Note, error) {
	if len(value) == 0 {
		return nil, fmt.Errorf("%w: note is required", exception.ErrInvalid)
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, value); err != nil {
		return nil, fmt.Errorf("%w: note must be valid JSON", exception.ErrInvalid)
	}
	compacted := buf.Bytes()

	if len(compacted) > NoteMaxBytes {
		return nil, fmt.Errorf("%w: note must be at most 100 kilobytes", exception.ErrInvalid)
	}

	var doc NoteNode
	if err := json.Unmarshal(compacted, &doc); err != nil || doc.Type != "doc" {
		return nil, fmt.Errorf("%w: note must be a document with type \"doc\"", exception.ErrInvalid)
	}

	return Note(compacted), nil
}

type NoteNode struct {
	Type    string     `json:"type"`
	Text    string     `json:"text"`
	Content []NoteNode `json:"content"`
}
