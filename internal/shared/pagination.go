package shared

type PageParams struct {
	Page     int
	PageSize int
}

func NormalizePage(page, pageSize int) PageParams {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return PageParams{Page: page, PageSize: pageSize}
}

func (p PageParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

type PageMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

func NewPageMeta(page PageParams, total int64) PageMeta {
	totalPages := int64(0)
	if page.PageSize > 0 {
		totalPages = (total + int64(page.PageSize) - 1) / int64(page.PageSize)
	}
	return PageMeta{
		Page:       page.Page,
		PageSize:   page.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}
