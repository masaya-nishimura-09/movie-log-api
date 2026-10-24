package record

import (
	"fmt"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type ArticleQuery struct {
	RecordID  ID
	Keyword   ArticleKeyword
	SortField ArticleSortField
	SortOrder SortOrder
	Page      Page
	PerPage   PerPage
}

func NewArticleQuery(
	recordID ID,
	keyword ArticleKeyword,
	sortField ArticleSortField,
	sortOrder SortOrder,
	page Page,
	perPage PerPage,
) ArticleQuery {
	return ArticleQuery{
		RecordID:  recordID,
		Keyword:   keyword,
		SortField: sortField,
		SortOrder: sortOrder,
		Page:      page,
		PerPage:   perPage,
	}
}

type ArticleKeyword string

func NewArticleKeyword(value string) (ArticleKeyword, error) {
	if utf8.RuneCountInString(value) > 255 {
		return "", fmt.Errorf("%w: article keyword must be at most 255 characters", exception.ErrInvalid)
	}

	return ArticleKeyword(value), nil
}

type ArticleSortField string

const (
	ArticleSortFieldCreatedAt ArticleSortField = "created_at"
	ArticleSortFieldUpdatedAt ArticleSortField = "updated_at"
	ArticleSortFieldTitle     ArticleSortField = "title"
)

func NewArticleSortField(value string) (ArticleSortField, error) {
	if value == "" {
		return "", fmt.Errorf("%w: article sort field is required", exception.ErrInvalid)
	}

	switch sortField := ArticleSortField(value); sortField {
	case ArticleSortFieldCreatedAt,
		ArticleSortFieldUpdatedAt,
		ArticleSortFieldTitle:
		return sortField, nil
	default:
		return "", fmt.Errorf("%w: invalid article sort field", exception.ErrInvalid)
	}
}
