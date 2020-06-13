package shared

import "math"

// Result common output
type Result struct {
	Data  interface{}
	Error error
}

// SliceResult include meta
type SliceResult struct {
	Data interface{}
	Meta Meta
}

// Meta model
type Meta struct {
	Page         int `json:"page"`
	Limit        int `json:"limit"`
	TotalRecords int `json:"totalRecords"`
	TotalPages   int `json:"totalPages"`
}

// NewMeta create new meta for slice data
func NewMeta(page, limit, totalRecords int) *Meta {
	var m Meta
	m.Page, m.Limit, m.TotalRecords = page, limit, totalRecords
	m.CalculatePages()
	return &m
}

// CalculatePages meta method
func (m *Meta) CalculatePages() {
	m.TotalPages = int(math.Ceil(float64(m.TotalRecords) / float64(m.Limit)))
}

// ToResolver graphql
func (m *Meta) ToResolver() *MetaResolver {
	return &MetaResolver{
		Page: int32(m.Page), Limit: int32(m.Limit), TotalRecords: int32(m.TotalRecords), TotalPages: int32(m.TotalPages),
	}
}

// MetaResolver model for graphql resolver, graphql doesn't support int64 data type (https://github.com/graphql/graphql-spec/issues/73)
type MetaResolver struct {
	Page         int32 `json:"page"`
	Limit        int32 `json:"limit"`
	TotalRecords int32 `json:"totalRecords"`
	TotalPages   int32 `json:"totalPages"`
}
