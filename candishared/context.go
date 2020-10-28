package candishared

import (
	"context"
)

// ContextKey represent Key of all context
type ContextKey string

// SetToContext will set context with specific key
func SetToContext(ctx context.Context, key ContextKey, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetValueFromContext will get context with specific key
func GetValueFromContext(ctx context.Context, key ContextKey) interface{} {
	return ctx.Value(key)
}

type contextKeyStruct struct{}

var (
	// HTTPHeaderContextKey context key
	HTTPHeaderContextKey = contextKeyStruct{}

	// TaskQueueRetryContextKey context key
	TaskQueueRetryContextKey = contextKeyStruct{}
)
