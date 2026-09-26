package record

import (
	"fmt"
	"slices"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Score uint

func NewScore(value uint) (Score, error) {
	if value < 1 || value > 5 {
		return 0, fmt.Errorf("%w: score must be between 1 and 5", exception.ErrInvalid)
	}

	return Score(value), nil
}

func NewScores(values []uint) ([]Score, error) {
	scores := make([]Score, 0, len(values))

	for _, value := range values {
		score, err := NewScore(value)
		if err != nil {
			return nil, err
		}

		if slices.Contains(scores, score) {
			return nil, fmt.Errorf("%w: duplicate score", exception.ErrInvalid)
		}

		scores = append(scores, score)
	}

	return scores, nil
}
