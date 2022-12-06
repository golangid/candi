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

		authorization := req.Header.Get(candihelper.HeaderAuthorization)
		if authorization == "" {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(w)
			return
		}

		authValues := strings.Split(authorization, " ")

		if len(authValues) != 2 {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(w)
			return
		}

		authType := strings.ToLower(authValues[0])

		tokenString := authValues[1]
		claimData, err := m.checkMultipleAuth(ctx, authType, tokenString)
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

// GraphQLAuth for graphql resolver
func (m *Middleware) GraphQLAuth(ctx context.Context, directive *gqltypes.Directive, input interface{}) (context.Context, error) {
	trace := tracer.StartTrace(ctx, "Middleware:GraphQLAuthDirective")
	defer trace.Finish()

	headers := ctx.Value(candishared.ContextKeyHTTPHeader).(http.Header)
	authorization := headers.Get(candihelper.HeaderAuthorization)
	trace.Log(candihelper.HeaderAuthorization, authorization)
	trace.SetTag("directiveName", directive.Name.Name)

	authValues := strings.Split(authorization, " ")
	if len(authValues) != 2 {
		return ctx, candishared.NewGraphQLErrorResolver("Invalid authorization", map[string]interface{}{
			"code":    401,
			"success": false,
		})
	}

	authTypeValue := directive.Arguments.MustGet("authType")
	if authTypeValue == nil {
		return ctx, candishared.NewGraphQLErrorResolver(
			"Missing authType argument in directive @"+directive.Name.Name+" definition",
			map[string]interface{}{
				"code":    401,
				"success": false,
			})
	}
	authType := strings.ToLower(authTypeValue.String())
	if authType != strings.ToLower(authValues[0]) {
		return ctx, candishared.NewGraphQLErrorResolver(
			"Mismatch authType definition from directive @"+directive.Name.Name+" (required: "+authType+", given: "+strings.ToLower(authValues[0])+")",
			map[string]interface{}{
				"code":    401,
				"success": false,
			})
	}

	tokenString := authValues[1]

	claimData, err := m.checkMultipleAuth(trace.Context(), authType, tokenString)
	if err != nil {
		return ctx, candishared.NewGraphQLErrorResolver(err.Error(), map[string]interface{}{
			"code":    401,
			"success": false,
		})
	}

	if claimData != nil {
		ctx = candishared.SetToContext(ctx, candishared.ContextKeyTokenClaim, claimData)
		trace.Log("token_claim", claimData)
	}

	return ctx, nil
}

func (m *Middleware) checkMultipleAuth(ctx context.Context, authType, token string) (claimData *candishared.TokenClaim, err error) {

	switch authType {
	case Bearer:
		claimData, err = m.Bearer(ctx, token)
	case Basic:
		err = m.Basic(ctx, token)
	default:
		return nil, errors.New("Invalid authorization type, must BEARER or BASIC")
	}

	return claimData, err
}
