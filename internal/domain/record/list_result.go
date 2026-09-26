package record

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type ListResult struct {
	Records    []*Record
	TotalCount TotalCount
}

func NewListResult(records []*Record, totalCount TotalCount) ListResult {
	return ListResult{
		Records:    records,
		TotalCount: totalCount,
	}
}

type TotalCount uint

func NewTotalCount(value int) (TotalCount, error) {
	if value < 0 {
		return 0, fmt.Errorf("%w: total count must not be negative", exception.ErrInvalid)
	}

	return TotalCount(value), nil
}
