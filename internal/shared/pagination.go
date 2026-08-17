package shared

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PageParams struct {
	Page     int
	PageSize int
}

type PageMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func NormalizePage(page, pageSize int) PageParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return PageParams{Page: page, PageSize: pageSize}
}

func (p PageParams) Offset() int { return (p.Page - 1) * p.PageSize }

func NewPageMeta(p PageParams, total int64) PageMeta {
	pages := total / int64(p.PageSize)
	if total%int64(p.PageSize) != 0 {
		pages++
	}
	return PageMeta{Page: p.Page, PageSize: p.PageSize, Total: total, TotalPages: pages}
}
