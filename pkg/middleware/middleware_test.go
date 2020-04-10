package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"agungdwiprasetyo.com/backend-microservices/config"
)

func TestNewMiddleware(t *testing.T) {
	mw := NewMiddleware(&config.Config{})
	assert.NotNil(t, mw)
}
