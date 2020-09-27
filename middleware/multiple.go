package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"pkg.agungdwiprasetyo.com/gendon/helper"
	"pkg.agungdwiprasetyo.com/gendon/wrapper"
)

// HTTPMultipleAuth mix basic & bearer auth
func (m *Middleware) HTTPMultipleAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					wrapper.NewHTTPResponse(http.StatusInternalServerError, fmt.Sprint(r)).JSON(c.Response())
				}
			}()

			// get auth
			authorization := c.Request().Header.Get(echo.HeaderAuthorization)
			if authorization == "" {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(c.Response())
			}

			// get auth type
			authValues := strings.Split(authorization, " ")

			// validate value
			if len(authValues) != 2 {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(c.Response())
			}

			authType := strings.ToLower(authValues[0])

			// set token
			tokenString := authValues[1]

			checkerFunc, ok := m.authTypeCheckerFunc[authType]
			if !ok {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization type").JSON(c.Response())
			}

			claimData, err := checkerFunc(c.Request().Context(), tokenString)
			if err != nil {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
			}

			if claimData != nil {
				c.Set(helper.TokenClaimKey, claimData)
			}

			return next(c)
		}
	}
}
