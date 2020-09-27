package wrapper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestCustomHTTPErrorHandler(t *testing.T) {
	e := echo.New()
	url := "/testing"
	req, err := http.NewRequest(echo.GET, url, nil)
	assert.NoError(t, err)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err = &echo.HTTPError{
		Code: http.StatusNotFound, Message: "Not Found",
	}
	CustomHTTPErrorHandler(err, c)
}
