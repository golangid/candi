package candishared

// Filter data
type Filter struct {
	Limit            int    `json:"limit" default:"10"`
	Page             int    `json:"page" default:"1"`
	Offset           int    `json:"-"`
	Search           string `json:"search,omitempty"`
	OrderBy          string `json:"orderBy,omitempty"`
	Sort             string `json:"sort,omitempty" default:"desc" lower:"true"`
	ShowAll          bool   `json:"showAll"`
	AllowEmptyFilter bool   `json:"-"`
}

// CalculateOffset method
func (f *Filter) CalculateOffset() int {
	f.Offset = (f.Page - 1) * f.Limit
	return f.Offset
}

// GetPage method
func (f *Filter) GetPage() int {
	return f.Page
}

// IncrPage method
func (f *Filter) IncrPage() {
	f.Page++
}

// GetLimit method
func (f *Filter) GetLimit() int {
	return f.Limit
}

// NullableFilter filter contains nullable value
type NullableFilter struct {
	Limit   *int
	Page    *int
	Search  *string
	Sort    *string
	ShowAll *bool
	OrderBy *string
}

func (n *NullableFilter) ToFilter() (filter Filter) {
	if n.Search != nil {
		filter.Search = *n.Search
	}
	if n.OrderBy != nil {
		filter.OrderBy = *n.OrderBy
	}
	if n.Sort != nil {
		filter.Sort = *n.Sort
	}
	if n.ShowAll != nil {
		filter.ShowAll = *n.ShowAll
	}

	if n.Limit == nil {
		filter.Limit = 10
	} else {
		filter.Limit = *n.Limit
	}
	if n.Page == nil {
		filter.Page = 1
	} else {
		filter.Page = *n.Page
	}
	return filter
}
