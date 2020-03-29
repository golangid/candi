package middleware

import (
	"net/http"
	"strings"

	"github.com/agungdwiprasetyo/backend-microservices/pkg/helper"
	"github.com/agungdwiprasetyo/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// ValidateBearer jwt token middleware
func (m *mw) ValidateBearer() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			authorization := c.Request().Header.Get(echo.HeaderAuthorization)
			if authorization == "" {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(c.Response())
			}

			authValues := strings.Split(authorization, " ")
			authType := strings.ToLower(authValues[0])
			if authType != "bearer" || len(authValues) != 2 {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(c.Response())
			}

			tokenString := authValues[1]
			resp := <-m.tokenUtil.Validate(c.Request().Context(), tokenString)
			if resp.Error != nil {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, resp.Error.Error()).JSON(c.Response())
			}

			c.Set(helper.TokenClaimKey, resp.Data)
			return next(c)
		}
	}
}
