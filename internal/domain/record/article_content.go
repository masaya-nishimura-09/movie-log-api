package record

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type ArticleContent json.RawMessage

const ArticleContentMaxBytes = 100 * 1024

func NewArticleContent(value json.RawMessage) (ArticleContent, error) {
	if len(value) == 0 {
		return nil, fmt.Errorf("%w: article content is required", exception.ErrInvalid)
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, value); err != nil {
		return nil, fmt.Errorf("%w: article content must be valid JSON", exception.ErrInvalid)
	}
	compacted := buf.Bytes()

	if len(compacted) > ArticleContentMaxBytes {
		return nil, fmt.Errorf("%w: article content must be at most 100 kilobytes", exception.ErrInvalid)
	}

	var doc ArticleContentNode
	if err := json.Unmarshal(compacted, &doc); err != nil || doc.Type != "doc" {
		return nil, fmt.Errorf("%w: article content must be a document with type \"doc\"", exception.ErrInvalid)
	}

	return ArticleContent(compacted), nil
}

type ArticleContentNode struct {
	Type    string               `json:"type"`
	Text    string               `json:"text"`
	Content []ArticleContentNode `json:"content"`
}
