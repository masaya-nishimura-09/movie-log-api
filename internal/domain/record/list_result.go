package record

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

type TotalCount uint
