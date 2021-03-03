package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMiddleware(t *testing.T) {
	mw := NewMiddleware(nil, nil)
	assert.NotNil(t, mw)
}
