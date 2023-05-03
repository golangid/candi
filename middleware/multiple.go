package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/candishared"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
	gqltypes "github.com/golangid/graphql-go/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// HTTPMultipleAuth mix basic & bearer auth
func (m *Middleware) HTTPMultipleAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				wrapper.NewHTTPResponse(http.StatusInternalServerError, fmt.Sprint(r)).JSON(w)
			}
		}()

		ctx := req.Context()

		trace := tracer.StartTrace(ctx, "Middleware:HTTPMultipleAuth")
		defer trace.Finish()

		authorization := req.Header.Get(candihelper.HeaderAuthorization)
		if authorization == "" {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(w)
			return
		}

		authType, authVal, ok := strings.Cut(authorization, " ")
		if !ok {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(w)
			return
		}
		claimData, err := m.checkMultipleAuth(trace.Context(), authType, authVal)
		if err != nil {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(w)
			return
		}

		if claimData != nil {
			ctx = candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, claimData)
		}

		next.ServeHTTP(w, req.WithContext(ctx))
	})
}

// GRPCMultipleAuth method
func (m *Middleware) GRPCMultipleAuth(ctx context.Context) (context.Context, error) {
	trace := tracer.StartTrace(ctx, "Middleware:GRPCMultipleAuth")
	defer trace.Finish()

	auth, err := extractAuthorizationGRPCMetadata(ctx)
	if err != nil {
		trace.SetError(err)
		return ctx, err
	}
	trace.Log(candihelper.HeaderAuthorization, auth)

	authType, authVal, ok := strings.Cut(auth, " ")
	if !ok {
		return ctx, grpc.Errorf(codes.Unauthenticated, "Invalid authorization")
	}
	claimData, err := m.checkMultipleAuth(trace.Context(), authType, authVal)
	if err != nil {
		return ctx, grpc.Errorf(codes.Unauthenticated, err.Error())
	}

	if claimData != nil {
		ctx = candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, claimData)
		trace.Log("token_claim", claimData)
	}
	return ctx, nil
}

// GraphQLAuth for graphql resolver
func (m *Middleware) GraphQLAuth(ctx context.Context, directive *gqltypes.Directive, input interface{}) (context.Context, error) {
	trace := tracer.StartTrace(ctx, "Middleware:GraphQLAuthDirective")
	defer trace.Finish()

	headers := ctx.Value(candishared.ContextKeyHTTPHeader).(http.Header)
	authorization := headers.Get(candihelper.HeaderAuthorization)
	trace.Log(candihelper.HeaderAuthorization, authorization)
	trace.SetTag("directiveName", directive.Name.Name)

	headerAuthType, headerAuthVal, ok := strings.Cut(authorization, " ")
	if !ok {
		return ctx, candishared.NewGraphQLErrorResolver("Invalid authorization", map[string]interface{}{
			"code": 401, "success": false,
		})
	}

	authTypeValue := directive.Arguments.MustGet("authType")
	if authTypeValue == nil {
		return ctx, candishared.NewGraphQLErrorResolver(
			"Missing authType argument in directive @"+directive.Name.Name+" definition",
			map[string]interface{}{"code": 401, "success": false},
		)
	}

	authType := authTypeValue.String()
	if _, ok := map[string]struct{}{BEARER: {}, BASIC: {}, MULTIPLE: {}}[authType]; !ok {
		return ctx, candishared.NewGraphQLErrorResolver(
			"Invalid authType direction name. Must BASIC, BEARER, or MULTIPLE",
			map[string]interface{}{"code": 401, "success": false},
		)
	}

	if authType != MULTIPLE && authType != strings.ToUpper(headerAuthType) {
		return ctx, candishared.NewGraphQLErrorResolver(
			"Mismatch authType definition from directive @"+directive.Name.Name+" (required: "+authType+", given: "+strings.ToUpper(headerAuthType)+")",
			map[string]interface{}{"code": 401, "success": false},
		)
	}

	claimData, err := m.checkMultipleAuth(trace.Context(), headerAuthType, headerAuthVal)
	if err != nil {
		return ctx, candishared.NewGraphQLErrorResolver(err.Error(), map[string]interface{}{
			"code": 401, "success": false,
		})
	}

	if claimData != nil {
		ctx = candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, claimData)
		trace.Log("token_claim", claimData)
	}

	return ctx, nil
}

func (m *Middleware) checkMultipleAuth(ctx context.Context, authType, token string) (claimData *candishared.TokenClaim, err error) {

	switch strings.ToUpper(authType) {
	case BEARER:
		claimData, err = m.Bearer(ctx, token)
	case BASIC:
		err = m.Basic(ctx, token)
	default:
		return nil, errors.New("Invalid authorization type, must BEARER or BASIC")
	}

	return claimData, err
}
