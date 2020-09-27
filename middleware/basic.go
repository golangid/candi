package middleware

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"pkg.agungdwiprasetyo.com/gendon/shared"
	"pkg.agungdwiprasetyo.com/gendon/wrapper"

	"github.com/labstack/echo"
)

const (
	// Basic constanta
	Basic = "basic"
)

// Basic function basic auth
func (m *Middleware) Basic(ctx context.Context, key string) error {

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
func (m *Middleware) HTTPBasicAuth(showAlert bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {

			if showAlert {
				c.Response().Header().Set(echo.HeaderWWWAuthenticate, `Basic realm=""`)
			}

			authorization := c.Request().Header.Get(echo.HeaderAuthorization)
			if authorization == "" {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(c.Response())
			}

			key, err := extractAuthType(Basic, authorization)
			if err != nil {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
			}

			if err := m.Basic(c.Request().Context(), key); err != nil {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
			}

			return next(c)
		}
	}
}

// GraphQLBasicAuth for graphql resolver
func (m *Middleware) GraphQLBasicAuth(ctx context.Context) {
	headers := ctx.Value(shared.ContextKey("headers")).(http.Header)
	authorization := headers.Get(echo.HeaderAuthorization)

	key, err := extractAuthType(Basic, authorization)
	if err != nil {
		panic(err)
	}

	if err := m.Basic(ctx, key); err != nil {
		panic(err)
	}
}

// GRPCBasicAuth method
func (m *Middleware) GRPCBasicAuth(ctx context.Context) {

	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		panic("missing context metadata")
	}

	authorizationMap := meta["authorization"]
	if len(authorizationMap) != 1 {
		panic(grpc.Errorf(codes.Unauthenticated, "Invalid authorization"))
	}

	authorization := authorizationMap[0]
	key, err := extractAuthType(Basic, authorization)
	if err != nil {
		panic(err)
	}

	if err := m.Basic(ctx, key); err != nil {
		panic(err)
	}
}
