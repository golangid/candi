package middleware

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
)

// Basic function basic auth
func (m *Middleware) Basic(ctx context.Context, key string) error {
	if m.basicAuthValidator == nil {
		return errors.New("Missing basic auth implementor")
	}

	data, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return errors.New("Unauthorized")
	}

	decoded := strings.Split(string(data), ":")
	if len(decoded) < 2 {
		return errors.New("Unauthorized")
	}
	username, password := decoded[0], decoded[1]
	return m.basicAuthValidator.ValidateBasic(ctx, username, password)
}

// HTTPBasicAuth http basic auth middleware
func (m *Middleware) HTTPBasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("WWW-Authenticate", `Basic realm=""`)

		if err := func(req *http.Request) error {
			trace := tracer.StartTrace(req.Context(), "Middleware:HTTPBasicAuth")
			defer trace.Finish()

			authorization := req.Header.Get(candihelper.HeaderAuthorization)
			trace.Log(candihelper.HeaderAuthorization, authorization)
			key, err := extractAuthType(BASIC, authorization)
			if err != nil {
				trace.SetError(err)
				return err
			}

			if err := m.Basic(req.Context(), key); err != nil {
				trace.SetError(err)
				return err
			}
			return nil
		}(req); err != nil {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(w)
			return
		}

		next.ServeHTTP(w, req)
	})
}

// GRPCBasicAuth method
func (m *Middleware) GRPCBasicAuth(ctx context.Context) (context.Context, error) {
	trace := tracer.StartTrace(ctx, "Middleware:GRPCBasicAuth")
	defer trace.Finish()

	auth, err := extractAuthorizationGRPCMetadata(ctx)
	if err != nil {
		trace.SetError(err)
		return ctx, err
	}
	trace.Log(candihelper.HeaderAuthorization, auth)

	key, err := extractAuthType(BASIC, auth)
	if err != nil {
		trace.SetError(err)
		return ctx, err
	}

	if err := m.Basic(trace.Context(), key); err != nil {
		trace.SetError(err)
		return ctx, err
	}

	return ctx, nil
}
