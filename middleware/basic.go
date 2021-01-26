package middleware

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	gqlerr "github.com/golangid/graphql-go/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"pkg.agungdwiprasetyo.com/candi/candishared"
	"pkg.agungdwiprasetyo.com/candi/config/env"
	"pkg.agungdwiprasetyo.com/candi/tracer"
	"pkg.agungdwiprasetyo.com/candi/wrapper"

	"github.com/labstack/echo"
)

const (
	// Basic constanta
	Basic = "basic"
)

// Basic function basic auth
func (m *Middleware) Basic(ctx context.Context, key string) error {
	if env.BaseEnv().NoAuth {
		return nil
	}

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
			if err := func() error {
				trace := tracer.StartTrace(c.Request().Context(), "Middleware:HTTPBasicAuth")
				defer trace.Finish()

				if showAlert {
					c.Response().Header().Set(echo.HeaderWWWAuthenticate, `Basic realm=""`)
				}

				authorization := c.Request().Header.Get(echo.HeaderAuthorization)
				key, err := extractAuthType(Basic, authorization)
				if err != nil {
					trace.SetError(err)
					return err
				}

				if err := m.Basic(c.Request().Context(), key); err != nil {
					trace.SetError(err)
					return err
				}
				return nil
			}(); err != nil {
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
			}

			return next(c)
		}
	}
}

// GraphQLBasicAuth for graphql resolver
func (m *Middleware) GraphQLBasicAuth(ctx context.Context) context.Context {
	trace := tracer.StartTrace(ctx, "Middleware:GraphQLBasicAuth")
	defer trace.Finish()

	headers := ctx.Value(candishared.ContextKeyHTTPHeader).(http.Header)
	authorization := headers.Get(echo.HeaderAuthorization)

	key, err := extractAuthType(Basic, authorization)
	if err != nil {
		trace.SetError(err)
		panic(&gqlerr.QueryError{
			Message: err.Error(),
			Extensions: map[string]interface{}{
				"code":    401,
				"success": false,
			},
		})
	}

	if err := m.Basic(trace.Context(), key); err != nil {
		trace.SetError(err)
		panic(&gqlerr.QueryError{
			Message: err.Error(),
			Extensions: map[string]interface{}{
				"code":    401,
				"success": false,
			},
		})
	}
	return ctx
}

// GRPCBasicAuth method
func (m *Middleware) GRPCBasicAuth(ctx context.Context) context.Context {
	trace := tracer.StartTrace(ctx, "Middleware:GRPCBasicAuth")
	defer trace.Finish()

	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		err := errors.New("missing context metadata")
		trace.SetError(err)
		panic(err)
	}

	authorizationMap := meta["authorization"]
	if len(authorizationMap) != 1 {
		err := grpc.Errorf(codes.Unauthenticated, "Invalid authorization")
		trace.SetError(err)
		panic(err)
	}

	key, err := extractAuthType(Basic, authorizationMap[0])
	if err != nil {
		trace.SetError(err)
		panic(err)
	}

	if err := m.Basic(trace.Context(), key); err != nil {
		trace.SetError(err)
		panic(err)
	}

	return ctx
}
