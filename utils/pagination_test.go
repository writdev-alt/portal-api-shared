package utils

import (
	"testing"
)

func TestPagination_Validate(t *testing.T) {
	tests := []struct {
		name     string
		p        Pagination
		expected Pagination
	}{
		{
			name:     "valid pagination",
			p:        Pagination{Page: 2, PerPage: 20},
			expected: Pagination{Page: 2, PerPage: 20},
		},
		{
			name:     "page too small",
			p:        Pagination{Page: 0, PerPage: 20},
			expected: Pagination{Page: 1, PerPage: 20},
		},
		{
			name:     "perPage too small",
			p:        Pagination{Page: 1, PerPage: 0},
			expected: Pagination{Page: 1, PerPage: 20},
		},
		{
			name:     "perPage too large",
			p:        Pagination{Page: 1, PerPage: 200},
			expected: Pagination{Page: 1, PerPage: 100},
		},
		{
			name:     "both invalid",
			p:        Pagination{Page: -1, PerPage: -1},
			expected: Pagination{Page: 1, PerPage: 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.Validate()

			if tt.p.Page != tt.expected.Page {
				t.Errorf("Page = %d, expected %d", tt.p.Page, tt.expected.Page)
			}

			if tt.p.PerPage != tt.expected.PerPage {
				t.Errorf("PerPage = %d, expected %d", tt.p.PerPage, tt.expected.PerPage)
			}
		})
	}
}

func TestPagination_Offset(t *testing.T) {
	tests := []struct {
		name     string
		p        Pagination
		expected int
	}{
		{
			name:     "page 1, perPage 20",
			p:        Pagination{Page: 1, PerPage: 20},
			expected: 0,
		},
		{
			name:     "page 2, perPage 20",
			p:        Pagination{Page: 2, PerPage: 20},
			expected: 20,
		},
		{
			name:     "page 3, perPage 10",
			p:        Pagination{Page: 3, PerPage: 10},
			expected: 20,
		},
		{
			name:     "page 5, perPage 15",
			p:        Pagination{Page: 5, PerPage: 15},
			expected: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.p.Offset()
			if result != tt.expected {
				t.Errorf("Offset() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestPagination_Limit(t *testing.T) {
	tests := []struct {
		name     string
		p        Pagination
		expected int
	}{
		{
			name:     "perPage 20",
			p:        Pagination{Page: 1, PerPage: 20},
			expected: 20,
		},
		{
			name:     "perPage 50",
			p:        Pagination{Page: 2, PerPage: 50},
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.p.Limit()
			if result != tt.expected {
				t.Errorf("Limit() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestNewPaginationInfo(t *testing.T) {
	tests := []struct {
		name     string
		p        *Pagination
		total    int64
		expected PaginationInfo
	}{
		{
			name:     "exact page count",
			p:        &Pagination{Page: 1, PerPage: 20},
			total:    100,
			expected: PaginationInfo{CurrentPage: 1, PerPage: 20, Total: 100, TotalPages: 5},
		},
		{
			name:     "remainder page",
			p:        &Pagination{Page: 1, PerPage: 20},
			total:    105,
			expected: PaginationInfo{CurrentPage: 1, PerPage: 20, Total: 105, TotalPages: 6},
		},
		{
			name:     "single page",
			p:        &Pagination{Page: 1, PerPage: 20},
			total:    15,
			expected: PaginationInfo{CurrentPage: 1, PerPage: 20, Total: 15, TotalPages: 1},
		},
		{
			name:     "empty result",
			p:        &Pagination{Page: 1, PerPage: 20},
			total:    0,
			expected: PaginationInfo{CurrentPage: 1, PerPage: 20, Total: 0, TotalPages: 1},
		},
		{
			name:     "last page",
			p:        &Pagination{Page: 5, PerPage: 20},
			total:    100,
			expected: PaginationInfo{CurrentPage: 5, PerPage: 20, Total: 100, TotalPages: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPaginationInfo(tt.p, tt.total)

			if result.CurrentPage != tt.expected.CurrentPage {
				t.Errorf("CurrentPage = %d, expected %d", result.CurrentPage, tt.expected.CurrentPage)
			}

			if result.PerPage != tt.expected.PerPage {
				t.Errorf("PerPage = %d, expected %d", result.PerPage, tt.expected.PerPage)
			}

			if result.Total != tt.expected.Total {
				t.Errorf("Total = %d, expected %d", result.Total, tt.expected.Total)
			}

			if result.TotalPages != tt.expected.TotalPages {
				t.Errorf("TotalPages = %d, expected %d", result.TotalPages, tt.expected.TotalPages)
			}
		})
	}
}

func BenchmarkPagination_Validate(b *testing.B) {
	p := Pagination{Page: 2, PerPage: 20}

	for i := 0; i < b.N; i++ {
		p.Validate()
	}
}

func BenchmarkPagination_Offset(b *testing.B) {
	p := Pagination{Page: 5, PerPage: 20}

	for i := 0; i < b.N; i++ {
		p.Offset()
	}
}
