package middleware

import (
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// BasicAuth function basic auth
func (m *Middleware) BasicAuth() echo.MiddlewareFunc {
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if m.username == username && m.password == password {
			return true, nil
		}
		return false, nil
	})
}
