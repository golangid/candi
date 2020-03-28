package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {

	midd := &Middleware{
		username: "user", password: "da1c25d8-37c8-41b1-afe2-42dd4825bfea",
	}

	t.Run("Test With Valid Auth", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Basic dXNlcjpkYTFjMjVkOC0zN2M4LTQxYjEtYWZlMi00MmRkNDgyNWJmZWE=")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := echo.HandlerFunc(func(c echo.Context) error {
			return c.JSON(http.StatusOK, c.String(http.StatusOK, "hello"))
		})

		mw := midd.BasicAuth()(handler)
		err := mw(c)
		assert.NoError(t, err)
	})

	t.Run("Test With Invalid Auth", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Basic MjIyMjphc2RzZA==")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := echo.HandlerFunc(func(c echo.Context) error {
			return c.JSON(http.StatusOK, c.String(http.StatusOK, "hello"))
		})

		mw := midd.BasicAuth()(handler)
		err := mw(c)
		assert.Error(t, err)
	})
}
