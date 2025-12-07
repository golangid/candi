package candishared

import (
	"net/url"
	"testing"

	"github.com/golangid/candi/candihelper"
	"github.com/stretchr/testify/assert"
)

func TestParseFromQueryParam(t *testing.T) {
	type Embed struct {
		Page        int       `json:"page"`
		Offset      int       `json:"-"`
		Sort        string    `json:"sort,omitempty" default:"desc" lower:"true"`
		Includes    []string  `json:"includes"`
		IncludeInts []int     `json:"include-ints"`
		Floats      []float64 `json:"floats"`
		Query       string    `query:"query"`
		NoTag       string
	}
	type params struct {
		Embed
		IsActive bool    `json:"isActive"`
		Ptr      *string `json:"ptr"`
	}

	t.Run("Testcase #1: Positive", func(t *testing.T) {
		urlVal, err := url.ParseQuery("page=1&ptr=val&isActive=true&floats=1.0,2.0&query=search&NoTag=notag&includes=one,two,three&include-ints=1,2,3")
		assert.NoError(t, err)

		var p params
		err = ParseFromQueryParam(urlVal, &p)
		assert.NoError(t, err)
		assert.Equal(t, p.Page, 1)
		assert.Equal(t, *p.Ptr, "val")
		assert.Equal(t, p.IsActive, true)
		assert.Equal(t, []string{"one", "two", "three"}, p.Includes)
		assert.Equal(t, []int{1, 2, 3}, p.IncludeInts)
		assert.Equal(t, []float64{1.0, 2.0}, p.Floats)
		assert.Equal(t, "search", p.Query)
		assert.Equal(t, "notag", p.NoTag)
	})
	t.Run("Testcase #2: Negative, invalid data type (string to int in struct)", func(t *testing.T) {
		urlVal, err := url.ParseQuery("page=undefined")
		assert.NoError(t, err)

		var p params
		err = ParseFromQueryParam(urlVal, &p)
		assert.Error(t, err)
	})
	t.Run("Testcase #3: Negative, invalid data type (not boolean)", func(t *testing.T) {
		urlVal, err := url.ParseQuery("isActive=terue")
		assert.NoError(t, err)

		var p params
		err = ParseFromQueryParam(urlVal, &p)
		assert.Error(t, err)
	})
	t.Run("Testcase #4: Negative, invalid target type (not pointer)", func(t *testing.T) {
		urlVal, err := url.ParseQuery("isActive=true")
		assert.NoError(t, err)

		var p params
		err = ParseFromQueryParam(urlVal, p)
		assert.Error(t, err)
	})
	t.Run("Testcase #5: Negative, invalid target type (not int slice)", func(t *testing.T) {
		urlVal, err := url.ParseQuery("include-ints=one,2,three&floats=one")
		assert.NoError(t, err)

		var p params
		err = ParseFromQueryParam(urlVal, &p)
		assert.Error(t, err)
	})
}

func TestParseToQueryParam(t *testing.T) {
	type VariantRequestParams struct {
		Filter      **string `json:"filter,omitempty"`
		FilterQuery string   `json:"filter[query],omitempty"`
		FilterSkuNo string   `json:"filter[skuNo],omitempty"`
		Page        int      `json:"page"`
		Limit       int      `json:"limit"`
		Ignore      string   `json:"-"`
	}

	var param VariantRequestParams
	param.Filter = candihelper.WrapPtr(candihelper.WrapPtr("product"))
	param.FilterQuery = "kulkas"
	param.FilterSkuNo = ""
	param.Page = 1
	param.Limit = 10

	want := "filter=product&filter[query]=kulkas&page=1&limit=10"
	assert.Equal(t, want, ParseToQueryParam(candihelper.WrapPtr(candihelper.WrapPtr(candihelper.WrapPtr(param)))))
}
