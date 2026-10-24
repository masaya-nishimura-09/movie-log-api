package record

type ArticleListResult struct {
	Articles      []ArticleSummary
	FilteredCount FilteredCount
	TotalCount    TotalCount
}

func NewArticleListResult(
	articles []ArticleSummary,
	filteredCount FilteredCount,
	totalCount TotalCount,
) ArticleListResult {
	return ArticleListResult{
		Articles:      articles,
		FilteredCount: filteredCount,
		TotalCount:    totalCount,
	}
}
