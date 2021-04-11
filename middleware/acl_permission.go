package middleware

import (
	"context"
	"net/http"

	gqlerr "github.com/golangid/graphql-go/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/codebase/factory/types"
	"pkg.agungdp.dev/candi/config/env"
	"pkg.agungdp.dev/candi/tracer"
	"pkg.agungdp.dev/candi/wrapper"
)

func (m *Middleware) checkACLPermissionFromContext(ctx context.Context, permissionCode string) (*candishared.TokenClaim, error) {
	tokenClaim := candishared.ParseTokenClaimFromContext(ctx)
	if env.BaseEnv().NoAuth {
		tokenClaim.Role = "GUEST"
		return tokenClaim, nil
	}

	role, err := m.ACLPermissionChecker.CheckPermission(ctx, tokenClaim.Subject, permissionCode)
	if err != nil {
		return tokenClaim, err
	}
	tokenClaim.Role = role
	return tokenClaim, nil
}

// HTTPPermissionACL http middleware for check acl permission
func (m *Middleware) HTTPPermissionACL(permissionCode string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := req.Context()
			if err := func(permCode string) (err error) {
				trace := tracer.StartTrace(ctx, "Middleware:HTTPPermissionACL")
				defer trace.Finish()
				trace.SetTag("permissionCode", permCode)

				tokenClaim, err := m.checkACLPermissionFromContext(trace.Context(), permCode)
				if err != nil {
					trace.SetError(err)
					return err
				}
				ctx = candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim)
				return nil
			}(permissionCode); err != nil {
				wrapper.NewHTTPResponse(http.StatusForbidden, err.Error()).JSON(w)
				return
			}

			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}

// GraphQLPermissionACL graphql resolver for check acl permission
func (m *Middleware) GraphQLPermissionACL(permissionCode string) types.MiddlewareFunc {
	return func(ctx context.Context) context.Context {
		trace := tracer.StartTrace(ctx, "Middleware:GraphQLPermissionACL")
		defer trace.Finish()
		trace.SetTag("permissionCode", permissionCode)

		tokenClaim, err := m.checkACLPermissionFromContext(trace.Context(), permissionCode)
		if err != nil {
			trace.SetError(err)
			panic(&gqlerr.QueryError{
				Message: err.Error(),
				Extensions: map[string]interface{}{
					"code":    403,
					"success": false,
				},
			})
		}
		return candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim)
	}
}

// GRPCPermissionACL grpc interceptor for check acl permission
func (m *Middleware) GRPCPermissionACL(permissionCode string) types.MiddlewareFunc {
	return func(ctx context.Context) context.Context {
		trace := tracer.StartTrace(ctx, "Middleware:GRPCPermissionACL")
		defer trace.Finish()
		trace.SetTag("permissionCode", permissionCode)

		tokenClaim, err := m.checkACLPermissionFromContext(trace.Context(), permissionCode)
		if err != nil {
			panic(grpc.Errorf(codes.PermissionDenied, err.Error()))
		}
		return candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim)
	}
}
