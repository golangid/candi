package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	gqlerr "github.com/golangid/graphql-go/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"pkg.agungdwiprasetyo.com/candi/candihelper"
	"pkg.agungdwiprasetyo.com/candi/candishared"
	"pkg.agungdwiprasetyo.com/candi/config/env"
	"pkg.agungdwiprasetyo.com/candi/tracer"
	"pkg.agungdwiprasetyo.com/candi/wrapper"

	"github.com/labstack/echo"
)

const (
	// Bearer constanta
	Bearer = "bearer"
)

// Bearer token validator
func (m *Middleware) Bearer(ctx context.Context, tokenString string) (*candishared.TokenClaim, error) {
	if env.BaseEnv().NoAuth {
		return &candishared.TokenClaim{
			StandardClaims: jwt.StandardClaims{
				Audience: "ANONYMOUS",
			},
		}, nil
	}

	tokenClaim, err := m.tokenValidator.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	return tokenClaim, nil
}

// HTTPBearerAuth http jwt token middleware
func (m *Middleware) HTTPBearerAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			trace := tracer.StartTrace(c.Request().Context(), "Middleware:HTTPBearerAuth")
			defer trace.Finish()

			authorization := c.Request().Header.Get(echo.HeaderAuthorization)
			tokenValue, err := extractAuthType(Bearer, authorization)
			if err != nil {
				trace.SetError(err)
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
			}

			tokenClaim, err := m.Bearer(trace.Context(), tokenValue)
			if err != nil {
				trace.SetError(err)
				return wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(c.Response())
			}

			c.Set(candihelper.TokenClaimKey, tokenClaim)
			return next(c)
		}
	}
}

// GraphQLBearerAuth for graphql resolver
func (m *Middleware) GraphQLBearerAuth(ctx context.Context) context.Context {
	trace := tracer.StartTrace(ctx, "Middleware:GraphQLBearerAuth")
	defer trace.Finish()
	tags := trace.Tags()

	headers := ctx.Value(candishared.ContextKeyHTTPHeader).(http.Header)
	authorization := headers.Get(echo.HeaderAuthorization)

	tokenValue, err := extractAuthType(Bearer, authorization)
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
	tags["token"] = tokenValue

	tokenClaim, err := m.Bearer(trace.Context(), tokenValue)
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

	tags["token_claim"] = tokenClaim
	return candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim)
}

// GRPCBearerAuth method
func (m *Middleware) GRPCBearerAuth(ctx context.Context) context.Context {
	trace := tracer.StartTrace(ctx, "Middleware:GRPCBearerAuth")
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

	tokenClaim, err := m.Bearer(ctx, authorizationMap[0])
	if err != nil {
		trace.SetError(err)
		panic(err)
	}

	return candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim)
}
