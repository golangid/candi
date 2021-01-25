package candishared

// Filter data
type Filter struct {
	Limit   int    `json:"limit" default:"10"`
	Page    int    `json:"page" default:"1"`
	Offset  int    `json:"-"`
	Search  string `json:"search,omitempty"`
	OrderBy string `json:"orderBy,omitempty"`
	Sort    string `json:"sort,omitempty" default:"desc" lower:"true"`
	ShowAll bool   `json:"showAll"`
}

// CalculateOffset method
func (f *Filter) CalculateOffset() int {
	f.Offset = (f.Page - 1) * f.Limit
	return f.Offset
}
