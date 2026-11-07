package record

type ArticleDetail struct {
	Article Article
	Record  ArticleRecordSummary
}

type ArticleRecordSummary struct {
	ID          ID
	Title       Title
	ReleaseYear ReleaseYear
	PosterURL   PosterURL
}
