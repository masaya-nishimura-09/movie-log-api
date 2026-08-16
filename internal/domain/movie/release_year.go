package movie

import (
	"fmt"
	"time"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type ReleaseYear uint

func NewReleaseYear(value uint) (ReleaseYear, error) {
	if value < 1888 {
		return 0, fmt.Errorf("%w: release year must be 1888 or later", exception.ErrInvalid)
	}

	if value > uint(time.Now().Year())+5 {
		return 0, fmt.Errorf("%w: release year must not be more than 5 years in the future", exception.ErrInvalid)
	}

	return ReleaseYear(value), nil
}
