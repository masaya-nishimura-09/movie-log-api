package record

import (
	"fmt"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

func NewMemo(value string) (Memo, error) {
	if utf8.RuneCountInString(value) > 1000 {
		return "", fmt.Errorf("%w: memo must be at most 1000 characters", exception.ErrInvalid)
	}

	return Memo(value), nil
}
