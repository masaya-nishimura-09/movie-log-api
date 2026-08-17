package record

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Runtime uint

func NewRuntime(value uint) (Runtime, error) {
	if value > 1440 {
		return 0, fmt.Errorf("%w: runtime must be at most 1440 minutes", exception.ErrInvalid)
	}

	return Runtime(value), nil
}
