package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
)

// Bearer token validator
func (m *Middleware) Bearer(ctx context.Context, tokenString string) (*candishared.TokenClaim, error) {
	if m.tokenValidator == nil {
		return nil, errors.New("Missing token validator")
	}

	tokenClaim, err := m.tokenValidator.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	return tokenClaim, nil
}

// HTTPBearerAuth http jwt token middleware
func (m *Middleware) HTTPBearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		claim, err := func(req *http.Request) (*candishared.TokenClaim, error) {
			trace := tracer.StartTrace(req.Context(), "Middleware:HTTPBearerAuth")
			defer trace.Finish()

			authorization := req.Header.Get(candihelper.HeaderAuthorization)
			trace.Log(candihelper.HeaderAuthorization, authorization)
			tokenValue, err := extractAuthType(BEARER, authorization)
			if err != nil {
				trace.SetError(err)
				return nil, err
			}

			tokenClaim, err := m.Bearer(trace.Context(), tokenValue)
			if err != nil {
				trace.SetError(err)
				return nil, err
			}
			trace.Log("token_claim", tokenClaim)

			return tokenClaim, nil
		}(req)
		if err != nil {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(w)
			return
		}

		next.ServeHTTP(w, req.WithContext(candishared.SetToContext(req.Context(), candishared.ContextKeyTokenClaim, claim)))
	})
}

// GRPCBearerAuth method
func (m *Middleware) GRPCBearerAuth(ctx context.Context) (context.Context, error) {
	trace := tracer.StartTrace(ctx, "Middleware:GRPCBearerAuth")
	defer trace.Finish()

	auth, err := extractAuthorizationGRPCMetadata(ctx)
	if err != nil {
		trace.SetError(err)
		return ctx, err
	}
	trace.Log(candihelper.HeaderAuthorization, auth)

	tokenClaim, err := m.Bearer(trace.Context(), auth)
	if err != nil {
		trace.SetError(err)
		return ctx, err
	}

	trace.Log("token_claim", tokenClaim)
	return candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim), nil
}
