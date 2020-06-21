package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/wrapper"
	"github.com/labstack/echo"
)

// Bearer token validator
func (m *mw) Bearer(ctx context.Context, tokenString string) (*shared.TokenClaim, error) {
	resp := <-m.tokenValidator.Validate(ctx, tokenString)
	if resp.Error != nil {
		return nil, resp.Error
	}

	tokenClaim, ok := resp.Data.(*shared.TokenClaim)
	if !ok {
		return nil, errors.New("Validate token: result is not claim data")
	}

	return tokenClaim, nil
}

// HTTPBearerAuth http jwt token middleware
func (m *mw) HTTPBearerAuth() echo.MiddlewareFunc {
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
			tokenClaim, err := m.Bearer(c.Request().Context(), tokenString)
			if err != nil {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
			}

			c.Set(helper.TokenClaimKey, tokenClaim)
			return next(c)
		}
	}
}

func (m *mw) GraphQLBearerAuth(ctx context.Context) *shared.TokenClaim {
	headers := ctx.Value(shared.ContextKey("headers")).(http.Header)
	authorization := headers.Get("Authorization")
	if authorization == "" {
		panic("Invalid authorization")
	}

	authValues := strings.Split(authorization, " ")
	authType := strings.ToLower(authValues[0])
	if authType != "bearer" || len(authValues) != 2 {
		panic("Invalid authorization")
	}

	tokenClaim, err := m.Bearer(ctx, authValues[1])
	if err != nil {
		panic(err)
	}

	return tokenClaim
}
