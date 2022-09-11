package middleware

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/golangid/candi/candihelper"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/tracer"
	"github.com/golangid/candi/wrapper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

		if username != env.BaseEnv().BasicAuthUsername || password != env.BaseEnv().BasicAuthPassword {
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
func (m *Middleware) HTTPBasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := func() error {
			trace := tracer.StartTrace(req.Context(), "Middleware:HTTPBasicAuth")
			defer trace.Finish()

			w.Header().Set("WWW-Authenticate", `Basic realm=""`)
			authorization := req.Header.Get(candihelper.HeaderAuthorization)
			trace.SetTag(candihelper.HeaderAuthorization, authorization)
			key, err := extractAuthType(Basic, authorization)
			if err != nil {
				trace.SetError(err)
				return err
			}

			if err := m.Basic(req.Context(), key); err != nil {
				trace.SetError(err)
				return err
			}
			return nil
		}(); err != nil {
			wrapper.NewHTTPResponse(http.StatusUnauthorized, err.Error()).JSON(w)
			return
		}

		next.ServeHTTP(w, req)
	})
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

	authorizationMap := meta[strings.ToLower(candihelper.HeaderAuthorization)]
	trace.SetTag(candihelper.HeaderAuthorization, authorizationMap)
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
