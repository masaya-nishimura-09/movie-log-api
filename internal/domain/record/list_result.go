package record

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type ListResult struct {
	Records       []*Record
	FilteredCount FilteredCount
	TotalCount    TotalCount
}

func NewListResult(records []*Record, filteredCount FilteredCount, totalCount TotalCount) ListResult {
	return ListResult{
		Records:       records,
		FilteredCount: filteredCount,
		TotalCount:    totalCount,
	}
}

type FilteredCount uint

func NewFilteredCount(value int) (FilteredCount, error) {
	if value < 0 {
		return 0, fmt.Errorf("%w: filtered count must not be negative", exception.ErrInvalid)
	}

	return FilteredCount(value), nil
}

type TotalCount uint

func NewTotalCount(value int) (TotalCount, error) {
	if value < 0 {
		return 0, fmt.Errorf("%w: total count must not be negative", exception.ErrInvalid)
	}

	return TotalCount(value), nil
}
