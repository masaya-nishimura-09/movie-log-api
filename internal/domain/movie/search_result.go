package movie

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type SearchResult struct {
	Movies       []*Movie
	Page         Page
	TotalPages   TotalPages
	TotalResults TotalResults
}

func NewSearchResult(
	movies []*Movie,
	page Page,
	totalPages TotalPages,
	totalResults TotalResults,
) SearchResult {
	return SearchResult{
		Movies:       movies,
		Page:         page,
		TotalPages:   totalPages,
		TotalResults: totalResults,
	}
}

type Page uint

func NewPage(value uint) (Page, error) {
	if value < 1 {
		return 0, fmt.Errorf("%w: page must be at least 1", exception.ErrInvalid)
	}

	return Page(value), nil
}

type TotalPages uint
type TotalResults uint
