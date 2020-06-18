package middleware

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicAuth(t *testing.T) {

	midd := &mw{
		username: "user", password: "da1c25d8-37c8-41b1-afe2-42dd4825bfea",
	}

	t.Run("Test With Valid Auth", func(t *testing.T) {

		err := midd.Basic(context.Background(), "Basic dXNlcjpkYTFjMjVkOC0zN2M4LTQxYjEtYWZlMi00MmRkNDgyNWJmZWE=")
		assert.NoError(t, err)
	})

	t.Run("Test With Invalid Auth #1", func(t *testing.T) {

		err := midd.Basic(context.Background(), "Basic MjIyMjphc2RzZA==")
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

		err := midd.Basic(context.Background(), "Basic zzzzzzz")
		assert.Error(t, err)
	})

	t.Run("Test With Invalid Auth #5", func(t *testing.T) {

		err := midd.Basic(context.Background(), "Basic dGVzdGluZw==")
		assert.Error(t, err)
	})
}
