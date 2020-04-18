package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// BasicAuth function basic auth
func (m *mw) BasicAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("WWW-Authenticate", `Basic realm=""`)
			authorizations := strings.Split(c.Request().Header.Get("Authorization"), " ")
			if len(authorizations) != 2 {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Unauthorized").JSON(c.Response())
			}

			authType, val := authorizations[0], authorizations[1]
			if authType != "Basic" {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Unauthorized").JSON(c.Response())
			}

			isValid := func() bool {
				data, err := base64.StdEncoding.DecodeString(val)
				if err != nil {
					return false
				}

				decoded := strings.Split(string(data), ":")
				if len(decoded) < 2 {
					return false
				}
				username, password := decoded[0], decoded[1]

				if username != m.username || password != m.password {
					return false
				}

				return true
			}

			if !isValid() {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Unauthorized").JSON(c.Response())
			}

			return next(c)
		}
	}
}
