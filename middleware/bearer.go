package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt"
	gqlerr "github.com/golangid/graphql-go/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/tracer"
	"pkg.agungdp.dev/candi/wrapper"
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
				Subject:  "USER_ID_DUMMY",
			},
		}, nil
	}

	tokenClaim, err := m.TokenValidator.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	return tokenClaim, nil
}

// HTTPBearerAuth http jwt token middleware
func (m *Middleware) HTTPBearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if err := func() error {
			trace := tracer.StartTrace(ctx, "Middleware:HTTPBearerAuth")
			defer trace.Finish()

			authorization := req.Header.Get("Authorization")
			trace.SetTag("authorization", authorization)
			tokenValue, err := extractAuthType(Bearer, authorization)
			if err != nil {
				trace.SetError(err)
				return err
			}

			tokenClaim, err := m.Bearer(trace.Context(), tokenValue)
			if err != nil {
				trace.SetError(err)
				return err
			}
			trace.Log("token_claim", tokenClaim)
			ctx = candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim)
			return nil
		}(); err != nil {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(w)
			return
		}

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// GraphQLBearerAuth for graphql resolver
func (m *Middleware) GraphQLBearerAuth(ctx context.Context) context.Context {
	trace := tracer.StartTrace(ctx, "Middleware:GraphQLBearerAuth")
	defer trace.Finish()

	headers := ctx.Value(candishared.ContextKeyHTTPHeader).(http.Header)
	authorization := headers.Get("Authorization")
	trace.SetTag("authorization", authorization)

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

	trace.Log("token_claim", tokenClaim)
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
	trace.SetTag("authorization", authorizationMap)
	if len(authorizationMap) != 1 {
		err := grpc.Errorf(codes.Unauthenticated, "Invalid authorization")
		trace.SetError(err)
		panic(err)
	}

	tokenClaim, err := m.Bearer(trace.Context(), authorizationMap[0])
	if err != nil {
		trace.SetError(err)
		panic(err)
	}

	trace.Log("token_claim", tokenClaim)
	return candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim)
}
