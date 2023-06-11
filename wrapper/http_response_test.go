package wrapper

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/stretchr/testify/assert"
)

func TestNewHTTPResponse(t *testing.T) {
	type Data struct {
		ID string `json:"id"`
	}

	multiError := candihelper.NewMultiError()
	multiError.Append("test", fmt.Errorf("error test"))

	type args struct {
		code    int
		message string
		params  []interface{}
	}
	tests := []struct {
		name string
		args args
		want *HTTPResponse
	}{
		{
			name: "Testcase #1: Response data list (include meta)",
			args: args{
				code:    http.StatusOK,
				message: "Fetch all data",
				params: []interface{}{
					[]Data{{ID: "061499700032"}, {ID: "061499700033"}},
					candishared.Meta{Page: 1, Limit: 10, TotalPages: 10, TotalRecords: 100},
				},
			},
			want: &HTTPResponse{
				Success: true,
				Code:    200,
				Message: "Fetch all data",
				Meta:    candishared.Meta{Page: 1, Limit: 10, TotalPages: 10, TotalRecords: 100},
				Data:    []Data{{ID: "061499700032"}, {ID: "061499700033"}},
			},
		},
		{
			name: "Testcase #2: Response data detail",
			args: args{
				code:    http.StatusOK,
				message: "Get detail data",
				params: []interface{}{
					Data{ID: "061499700032"},
				},
			},
			want: &HTTPResponse{
				Success: true,
				Code:    200,
				Message: "Get detail data",
				Data:    Data{ID: "061499700032"},
			},
		},
		{
			name: "Testcase #3: Response only message (without data)",
			args: args{
				code:    http.StatusOK,
				message: "list data empty",
			},
			want: &HTTPResponse{
				Success: true,
				Code:    200,
				Message: "list data empty",
			},
		},
		{
			name: "Testcase #4: Response failed (ex: Bad Request)",
			args: args{
				code:    http.StatusBadRequest,
				message: "id cannot be empty",
				params:  []interface{}{multiError},
			},
			want: &HTTPResponse{
				Success: false,
				Code:    400,
				Message: "id cannot be empty",
				Errors:  map[string]string{"test": "error test"},
			},
		},
		{
			name: "Testcase #5: Response failed (error detail)",
			args: args{
				code:    http.StatusBadRequest,
				message: "Failed validate",
				params:  []interface{}{errors.New("error")},
			},
			want: &HTTPResponse{
				Success: false,
				Code:    400,
				Message: "Failed validate",
				Errors:  map[string]string{"detail": "error"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewHTTPResponse(tt.args.code, tt.args.message, tt.args.params...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\x1b[31;1mNewHTTPResponse() = %v, \nwant => %v\x1b[0m", got, tt.want)
			}
		})
	}
}

func TestHTTPResponse_JSON(t *testing.T) {
	rec := httptest.NewRecorder()
	resp := NewHTTPResponse(200, "success")
	assert.NoError(t, resp.JSON(rec))
}

func TestHTTPResponse_XML(t *testing.T) {
	rec := httptest.NewRecorder()
	resp := NewHTTPResponse(200, "success")
	assert.NoError(t, resp.XML(rec))
}
