package record

import (
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

func NewWatchedAt(value time.Time) (time.Time, error) {
	if value.IsZero() {
		return time.Time{}, fmt.Errorf("%w: watched at is required", exception.ErrInvalid)
	}

	if value.After(time.Now()) {
		return time.Time{}, fmt.Errorf("%w: watched at must not be in the future", exception.ErrInvalid)
	}

	return value, nil
}
