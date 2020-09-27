package shared

// Filter data
type Filter struct {
	Limit   int32  `json:"limit" default:"10"`
	Page    int32  `json:"page" default:"1"`
	Offset  int32  `json:"-"`
	Search  string `json:"search,omitempty"`
	OrderBy string `json:"orderBy,omitempty"`
	Sort    string `json:"sort,omitempty" default:"desc" lower:"true"`
}

// CalculateOffset method
func (f *Filter) CalculateOffset() {
	f.Offset = (f.Page - 1) * f.Limit
}
