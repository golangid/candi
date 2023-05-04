package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golangid/candi/candihelper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

const (
	basicUsername  = "user"
	basicPass      = "da1c25d8-37c8-41b1-afe2-42dd4825bfea"
	validBasicAuth = "dXNlcjpkYTFjMjVkOC0zN2M4LTQxYjEtYWZlMi00MmRkNDgyNWJmZWE="
)

func TestBasicAuth(t *testing.T) {

	midd := &Middleware{
		basicAuthValidator: &defaultMiddleware{
			username: basicUsername, password: basicPass,
		},
	}

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

	mw := &Middleware{
		basicAuthValidator: &defaultMiddleware{
			username: basicUsername, password: basicPass,
		},
	}

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

func TestMiddleware_GRPCBasicAuth(t *testing.T) {

	mw := &Middleware{
		basicAuthValidator: &defaultMiddleware{
			username: basicUsername, password: basicPass,
		},
	}

	t.Run("Testcase #1: Positive", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
			strings.ToLower(candihelper.HeaderAuthorization): []string{"Basic " + validBasicAuth},
		})
		_, err := mw.GRPCBasicAuth(ctx)
		assert.NoError(t, err)
	})
	t.Run("Testcase #2: Negative", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
			strings.ToLower(candihelper.HeaderAuthorization): []string{},
		})
		_, err := mw.GRPCBasicAuth(ctx)
		assert.Error(t, err)
	})
	t.Run("Testcase #3: Negative", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
			strings.ToLower(candihelper.HeaderAuthorization): []string{"Basic xxx"},
		})
		_, err := mw.GRPCBasicAuth(ctx)
		assert.Error(t, err)
	})
}
