package record

import (
	"fmt"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Query struct {
	Scores       []Score
	Platforms    []Platform
	MoodTags     []MoodTag
	Genres       []Genre
	TitleKeyword TitleKeyword
	SortField    SortField
	SortOrder    SortOrder
	Page         Page
	PerPage      PerPage
}

func NewQuery(
	scores []Score,
	platforms []Platform,
	moodTags []MoodTag,
	genres []Genre,
	titleKeyword TitleKeyword,
	sortField SortField,
	sortOrder SortOrder,
	page Page,
	perPage PerPage,
) Query {
	return Query{
		Scores:       scores,
		Platforms:    platforms,
		MoodTags:     moodTags,
		Genres:       genres,
		TitleKeyword: titleKeyword,
		SortField:    sortField,
		SortOrder:    sortOrder,
		Page:         page,
		PerPage:      perPage,
	}
}

type TitleKeyword string

func NewTitleKeyword(value string) (TitleKeyword, error) {
	if utf8.RuneCountInString(value) > 255 {
		return "", fmt.Errorf("%w: title keyword must be at most 255 characters", exception.ErrInvalid)
	}

	return TitleKeyword(value), nil
}

type SortField string

const (
	SortFieldWatchedAt   SortField = "watched_at"
	SortFieldReleaseYear SortField = "release_year"
	SortFieldScore       SortField = "score"
	SortFieldTitle       SortField = "title"
)

func NewSortField(value string) (SortField, error) {
	if value == "" {
		return SortFieldWatchedAt, nil
	}

	switch sortField := SortField(value); sortField {
	case SortFieldWatchedAt,
		SortFieldReleaseYear,
		SortFieldScore,
		SortFieldTitle:
		return sortField, nil
	default:
		return "", fmt.Errorf("%w: invalid sort field", exception.ErrInvalid)
	}
}

type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)

func NewSortOrder(value string) (SortOrder, error) {
	if value == "" {
		return SortOrderDesc, nil
	}

	switch sortOrder := SortOrder(value); sortOrder {
	case SortOrderAsc,
		SortOrderDesc:
		return sortOrder, nil
	default:
		return "", fmt.Errorf("%w: invalid sort order", exception.ErrInvalid)
	}
}

type Page uint

func NewPage(value uint) (Page, error) {
	if value < 1 {
		return 0, fmt.Errorf("%w: page must be at least 1", exception.ErrInvalid)
	}

	return Page(value), nil
}

type PerPage uint

func NewPerPage(value uint) (PerPage, error) {
	if value < 1 || value > 100 {
		return 0, fmt.Errorf("%w: per page must be between 1 and 100", exception.ErrInvalid)
	}

	return PerPage(value), nil
}
