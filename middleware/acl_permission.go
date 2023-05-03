package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/codebase/factory/types"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
	gqltypes "github.com/golangid/graphql-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func (m *Middleware) checkACLPermissionFromContext(ctx context.Context, permissionCode string) (tokenClaim *candishared.TokenClaim, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = errors.New("Missing token claim in context")
		}
	}()

	tokenClaim = candishared.ParseTokenClaimFromContext(ctx)

	if m.aclPermissionChecker == nil {
		return tokenClaim, errors.New("Missing acl permission checker")
	}

	userID := tokenClaim.Subject
	if m.extractUserIDFunc != nil {
		userID = m.extractUserIDFunc(tokenClaim)
	}
	role, err := m.aclPermissionChecker.CheckPermission(ctx, userID, permissionCode)
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
func (m *Middleware) GraphQLPermissionACL(ctx context.Context, directive *gqltypes.Directive, input interface{}) (context.Context, error) {
	trace := tracer.StartTrace(ctx, "Middleware:GraphQLPermissionACL")
	defer trace.Finish()

	permissionCode := directive.Arguments.MustGet("permissionCode")
	if permissionCode == nil {
		return ctx, candishared.NewGraphQLErrorResolver(
			"Missing permissionCode argument in directive @"+directive.Name.Name+" definition",
			map[string]interface{}{
				"code":    403,
				"success": false,
			})
	}

	trace.SetTag("directiveName", directive.Name.Name)
	trace.SetTag("permissionCode", permissionCode.String())

	tokenClaim, err := m.checkACLPermissionFromContext(trace.Context(), permissionCode.String())
	if err != nil {
		trace.SetError(err)
		return ctx, candishared.NewGraphQLErrorResolver(
			err.Error(),
			map[string]interface{}{
				"code":    403,
				"success": false,
			})
	}
	return candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim), nil
}

// GRPCPermissionACL grpc interceptor for check acl permission
func (m *Middleware) GRPCPermissionACL(permissionCode string) types.MiddlewareFunc {
	return func(ctx context.Context) (context.Context, error) {
		trace := tracer.StartTrace(ctx, "Middleware:GRPCPermissionACL")
		defer trace.Finish()
		trace.SetTag("permissionCode", permissionCode)

		tokenClaim, err := m.checkACLPermissionFromContext(trace.Context(), permissionCode)
		if err != nil {
			return ctx, grpc.Errorf(codes.PermissionDenied, err.Error())
		}
		return candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, tokenClaim), nil
	}
}
