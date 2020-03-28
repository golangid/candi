package wrapper

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

// CustomHTTPErrorHandler custom echo http error
func CustomHTTPErrorHandler(err error, c echo.Context) {
	var message string
	code := http.StatusInternalServerError
	if err != nil {
		message = err.Error()
	}

	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if code == http.StatusNotFound {
			message = fmt.Sprintf(`Resource "%s %s" not found`, c.Request().Method, c.Request().URL.Path)
		}
	}
	NewHTTPResponse(code, message).JSON(c.Response())
}
