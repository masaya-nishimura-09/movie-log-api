package record

type ArticleListResult struct {
	Articles      []ArticleDetail
	FilteredCount FilteredCount
	TotalCount    TotalCount
}

func NewArticleListResult(
	articles []ArticleDetail,
	filteredCount FilteredCount,
	totalCount TotalCount,
) ArticleListResult {
	return ArticleListResult{
		Articles:      articles,
		FilteredCount: filteredCount,
		TotalCount:    totalCount,
	}
}
