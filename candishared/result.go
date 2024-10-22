package candishared

import "math"

// Result common output
type Result struct {
	Data  any
	Error error
}

// SliceResult include meta
type SliceResult struct {
	Data  any
	Meta  Meta
	Error error
}

// Meta model
type Meta struct {
	Page         int `json:"page"`
	Limit        int `json:"limit"`
	TotalRecords int `json:"totalRecords"`
	TotalPages   int `json:"totalPages"`
}

// NewMeta create new meta for slice data
func NewMeta(page, limit, totalRecords int) (m Meta) {
	m.Page, m.Limit, m.TotalRecords = page, limit, totalRecords
	m.CalculatePages()
	return m
}

// CalculatePages meta method
func (m *Meta) CalculatePages() {
	m.TotalPages = int(math.Ceil(float64(m.TotalRecords) / float64(m.Limit)))
}
