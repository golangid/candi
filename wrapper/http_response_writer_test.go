package wrapper

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapHTTPResponseWriter(t *testing.T) {
	buff := new(bytes.Buffer)
	httpResp := NewWrapHTTPResponseWriter(buff, &httptest.ResponseRecorder{HeaderMap: http.Header{"a": []string{"b"}}})
	httpResp.WriteHeader(200)
	httpResp.Write([]byte("test"))
	assert.Equal(t, 200, httpResp.StatusCode())
	assert.NotNil(t, httpResp.Header())
}
