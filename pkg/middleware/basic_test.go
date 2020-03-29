package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {

	midd := &mw{
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

	t.Run("Test With Invalid Auth #1", func(t *testing.T) {
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
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusUnauthorized)
	})

	t.Run("Test With Invalid Auth #2", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Basic")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := echo.HandlerFunc(func(c echo.Context) error {
			return c.JSON(http.StatusOK, c.String(http.StatusOK, "hello"))
		})

		mw := midd.BasicAuth()(handler)
		err := mw(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusUnauthorized)
	})

	t.Run("Test With Invalid Auth #3", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer xxxx")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := echo.HandlerFunc(func(c echo.Context) error {
			return c.JSON(http.StatusOK, c.String(http.StatusOK, "hello"))
		})

		mw := midd.BasicAuth()(handler)
		err := mw(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusUnauthorized)
	})

	t.Run("Test With Invalid Auth #4", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Basic zzzzzz")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := echo.HandlerFunc(func(c echo.Context) error {
			return c.JSON(http.StatusOK, c.String(http.StatusOK, "hello"))
		})

		mw := midd.BasicAuth()(handler)
		err := mw(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusUnauthorized)
	})

	t.Run("Test With Invalid Auth #5", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Basic dGVzdGluZw==")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := echo.HandlerFunc(func(c echo.Context) error {
			return c.JSON(http.StatusOK, c.String(http.StatusOK, "hello"))
		})

		mw := midd.BasicAuth()(handler)
		err := mw(c)
		assert.NoError(t, err)
		assert.Equal(t, rec.Code, http.StatusUnauthorized)
	})
}
