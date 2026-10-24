package record

import "time"

type ArticleSummary struct {
	ID        ArticleID
	Record    ArticleRecordSummary
	Title     ArticleTitle
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ArticleRecordSummary struct {
	ID          ID
	Title       Title
	ReleaseYear ReleaseYear
	PosterURL   PosterURL
}
