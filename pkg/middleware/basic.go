package middleware

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// Basic function basic auth
func (m *mw) Basic(ctx context.Context, key string) error {

	isValid := func() bool {
		data, err := base64.StdEncoding.DecodeString(key)
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
		return errors.New("Unauthorized")
	}

	return nil
}

// HTTPBasicAuth http basic auth middleware
func (m *mw) HTTPBasicAuth(showAlert bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if showAlert {
				c.Response().Header().Set("WWW-Authenticate", `Basic realm=""`)
			}

			authorization := c.Request().Header.Get(echo.HeaderAuthorization)
			if authorization == "" {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(c.Response())
			}

			authValues := strings.Split(authorization, " ")
			authType := strings.ToLower(authValues[0])
			if authType != "basic" || len(authValues) != 2 {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(c.Response())
			}

			key := authValues[1]
			if err := m.Basic(c.Request().Context(), key); err != nil {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
			}

			return next(c)
		}
	}
}

func (m *mw) GraphQLBasicAuth(ctx context.Context) {
	headers := ctx.Value(shared.ContextKey("headers")).(http.Header)
	authorization := headers.Get("Authorization")

	authValues := strings.Split(authorization, " ")
	authType := strings.ToLower(authValues[0])
	if authType != "basic" || len(authValues) != 2 {
		panic("Invalid authorization")
	}

	if err := m.Basic(ctx, authValues[1]); err != nil {
		panic(err)
	}
}
