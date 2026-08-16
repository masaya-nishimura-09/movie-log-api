package record

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type WatchCount uint

func NewWatchCount(value uint) (WatchCount, error) {
	if value < 1 {
		return 0, fmt.Errorf("%w: watch count must be 1 or more", exception.ErrInvalid)
	}

	return WatchCount(value), nil
}
