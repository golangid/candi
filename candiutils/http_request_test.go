package candiutils

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"pkg.agungdp.dev/candi/candishared"
)

func TestNewRequest(t *testing.T) {
	testName := "Test positive new http request"

	t.Run(testName, func(t *testing.T) {
		// set new request
		request := NewHTTPRequest(
			HTTPRequestSetRetries(1),
			HTTPRequestSetSleepBetweenRetry(500*time.Millisecond),
			HTTPRequestSetHTTPErrorCodeThreshold(http.StatusBadRequest),
		)

		assert.NotNil(t, request)
	})
}

func TestRequestDo(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// mock data
	urlMock := "http://agungdp.dev"
	headerMock := map[string]string{"Content-Type": "application/json"}
	successResponseMock := map[string]interface{}{"success": true, "message": "success", "code": http.StatusOK}
	errorResponseMock := map[string]interface{}{"success": false, "message": "error", "code": http.StatusBadGateway}

	testCase := map[string]struct {
		wantError bool
		url       string
		code      int
		method    string
		body      interface{}
		header    map[string]string
		response  interface{}
	}{
		"Test #1 positive http request do": {
			wantError: false,
			url:       urlMock,
			code:      http.StatusOK,
			method:    http.MethodPost,
			response:  successResponseMock,
			body:      nil,
			header:    headerMock,
		},
		"Test #2 negative http request do client request": {
			wantError: true,
			url:       urlMock,
			code:      http.StatusBadGateway,
			method:    http.MethodPut,
			response:  errorResponseMock,
			body: &candishared.Result{
				Data: false,
			},
			header: map[string]string{},
		},
	}

	for name, test := range testCase {
		t.Run(name, func(t *testing.T) {
			// set new request
			request := NewHTTPRequest(
				HTTPRequestSetRetries(1),
				HTTPRequestSetSleepBetweenRetry(500*time.Millisecond),
				HTTPRequestSetHTTPErrorCodeThreshold(http.StatusBadRequest),
				HTTPRequestSetTLS(nil),
				HTTPRequestSetTimeout(5*time.Second),
				HTTPRequestSetBreakerName("test"),
			)

			if test.code < 500 {
				httpmock.RegisterResponder(http.MethodPost, test.url, func(req *http.Request) (*http.Response, error) {
					resp, _ := httpmock.NewJsonResponse(test.code, test.response)
					return resp, nil
				})
			}

			var (
				err error
			)

			if test.body != nil {
				req, _ := json.Marshal(test.body)

				// do request
				_, _, err = request.Do(context.Background(), test.method, test.url, req, test.header)
			} else {
				// do request
				_, _, err = request.Do(context.Background(), test.method, test.url, nil, test.header)
			}

			if test.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
