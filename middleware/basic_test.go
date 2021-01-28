package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"pkg.agungdwiprasetyo.com/candi/candishared"
	"pkg.agungdwiprasetyo.com/candi/config/env"
)

const (
	basicUsername  = "user"
	basicPass      = "da1c25d8-37c8-41b1-afe2-42dd4825bfea"
	validBasicAuth = "dXNlcjpkYTFjMjVkOC0zN2M4LTQxYjEtYWZlMi00MmRkNDgyNWJmZWE="
)

func TestBasicAuth(t *testing.T) {

	env.SetEnv(env.Env{BasicAuthUsername: basicUsername, BasicAuthPassword: basicPass})
	midd := &Middleware{}

	t.Run("Test With Valid Auth", func(t *testing.T) {

		err := midd.Basic(context.Background(), validBasicAuth)
		assert.NoError(t, err)
	})

	t.Run("Test With Invalid Auth #1", func(t *testing.T) {

		err := midd.Basic(context.Background(), "MjIyMjphc2RzZA==")
		assert.Error(t, err)
	})

	t.Run("Test With Invalid Auth #2", func(t *testing.T) {

		err := midd.Basic(context.Background(), "Basic")
		assert.Error(t, err)
	})

	t.Run("Test With Invalid Auth #3", func(t *testing.T) {

		err := midd.Basic(context.Background(), "Bearer xxx")
		assert.Error(t, err)
	})

	t.Run("Test With Invalid Auth #4", func(t *testing.T) {

		err := midd.Basic(context.Background(), "zzzzzzz")
		assert.Error(t, err)
	})

	t.Run("Test With Invalid Auth #5", func(t *testing.T) {

		err := midd.Basic(context.Background(), "Basic dGVzdGluZw==")
		assert.Error(t, err)
	})
}

func TestMiddleware_HTTPBasicAuth(t *testing.T) {
	env.SetEnv(env.Env{BasicAuthUsername: basicUsername, BasicAuthPassword: basicPass})
	mw := &Middleware{}

	tests := []struct {
		name             string
		authorization    string
		wantResponseCode int
	}{
		{
			name:             "Testcase #1: Positive",
			authorization:    "Basic " + validBasicAuth,
			wantResponseCode: 200,
		},
		{
			name:             "Testcase #2: Negative, empty authorization",
			authorization:    "",
			wantResponseCode: 401,
		},
		{
			name:             "Testcase #3: Negative, invalid authorization",
			authorization:    "Basic xxxxx",
			wantResponseCode: 401,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

			handlerToTest := mw.HTTPBasicAuth(nextHandler)

			req := httptest.NewRequest("GET", "http://testing", nil)
			req.Header.Add("Authorization", tt.authorization)

			recorder := httptest.NewRecorder()
			handlerToTest.ServeHTTP(recorder, req)

			assert.Equal(t, tt.wantResponseCode, recorder.Code)
		})
	}
}

func TestMiddleware_GraphQLBasicAuth(t *testing.T) {
	env.SetEnv(env.Env{BasicAuthUsername: basicUsername, BasicAuthPassword: basicPass})
	mw := &Middleware{}

	t.Run("Testcase #1: Positive", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), candishared.ContextKeyHTTPHeader, http.Header{
			"Authorization": []string{"Basic " + validBasicAuth},
		})
		assert.NotPanics(t, func() { mw.GraphQLBasicAuth(ctx) })
	})
	t.Run("Testcase #2: Negative", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), candishared.ContextKeyHTTPHeader, http.Header{
			"Authorization": []string{},
		})
		assert.Panics(t, func() { mw.GraphQLBasicAuth(ctx) })
	})
	t.Run("Testcase #3: Negative", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), candishared.ContextKeyHTTPHeader, http.Header{
			"Authorization": []string{"Basic xxx"},
		})
		assert.Panics(t, func() { mw.GraphQLBasicAuth(ctx) })
	})
}

func TestMiddleware_GRPCBasicAuth(t *testing.T) {
	env.SetEnv(env.Env{BasicAuthUsername: basicUsername, BasicAuthPassword: basicPass})
	mw := &Middleware{}

	t.Run("Testcase #1: Positive", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
			"authorization": []string{"Basic " + validBasicAuth},
		})
		assert.NotPanics(t, func() { mw.GRPCBasicAuth(ctx) })
	})
	t.Run("Testcase #2: Negative", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
			"authorization": []string{},
		})
		assert.Panics(t, func() { mw.GRPCBasicAuth(ctx) })
	})
	t.Run("Testcase #3: Negative", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
			"authorization": []string{"Basic xxx"},
		})
		assert.Panics(t, func() { mw.GRPCBasicAuth(ctx) })
	})
}
