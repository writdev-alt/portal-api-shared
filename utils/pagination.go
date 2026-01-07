package utils

// Pagination represents pagination parameters
type Pagination struct {
	Page    int `form:"page,default=1" json:"page,omitempty"`
	PerPage int `form:"per_page,default=20" json:"per_page,omitempty"`
}

// Validate validates and normalizes pagination parameters
func (p *Pagination) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

// Offset returns the offset for database query
func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the limit for database query
func (p *Pagination) Limit() int {
	return p.PerPage
}

// PaginationInfo represents pagination metadata in response
type PaginationInfo struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
}

// NewPaginationInfo creates pagination info from pagination and total count
func NewPaginationInfo(p *Pagination, total int64) PaginationInfo {
	totalPages := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))
	if totalPages < 1 {
		totalPages = 1
	}

	return PaginationInfo{
		CurrentPage: p.Page,
		PerPage:     p.PerPage,
		Total:       total,
		TotalPages:  totalPages,
	}
}
