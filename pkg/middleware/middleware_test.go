package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/agungdwiprasetyo/backend-microservices/config"
)

func TestNewMiddleware(t *testing.T) {
	mw := NewMiddleware(&config.Config{})
	assert.NotNil(t, mw)
}
