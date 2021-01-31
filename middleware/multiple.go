package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"pkg.agungdp.dev/candi/candishared"
	"pkg.agungdp.dev/candi/wrapper"
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

		// get auth
		authorization := req.Header.Get(echo.HeaderAuthorization)
		if authorization == "" {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(w)
			return
		}

		// get auth type
		authValues := strings.Split(authorization, " ")

		// validate value
		if len(authValues) != 2 {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization").JSON(w)
			return
		}

		authType := strings.ToLower(authValues[0])

		// set token
		tokenString := authValues[1]

		checkerFunc, ok := m.authTypeCheckerFunc[authType]
		if !ok {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, "Invalid authorization type").JSON(w)
			return
		}

		claimData, err := checkerFunc(ctx, tokenString)
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
