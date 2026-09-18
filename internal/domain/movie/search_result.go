package movie

type SearchResult struct {
	Movies       []*Movie
	Page         Page
	TotalPages   TotalPages
	TotalResults TotalResults
}

type Page uint
type TotalPages uint
type TotalResults uint
