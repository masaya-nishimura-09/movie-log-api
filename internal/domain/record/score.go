package record

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

func NewScore(value uint) (Score, error) {
	if value < 1 || value > 5 {
		return 0, fmt.Errorf("%w: score must be between 1 and 5", exception.ErrInvalid)
	}

	return Score(value), nil
}
