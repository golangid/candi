package candishared

import "context"

// ContextKey represent Key of all context
type ContextKey string

const (
	// ContextKeyHTTPHeader context key
	ContextKeyHTTPHeader ContextKey = "httpHeader"

	// ContextKeyTaskQueueRetry context key
	ContextKeyTaskQueueRetry ContextKey = "taskQueueRetry"

	// ContextKeyTokenClaim context key
	ContextKeyTokenClaim ContextKey = "tokenClaim"

	// ContextKeyWorkerKey context key
	ContextKeyWorkerKey ContextKey = "workerKey"

	// ContextKeySQLTransaction context key
	ContextKeySQLTransaction ContextKey = "sqltx"
)

// SetToContext will set context with specific key
func SetToContext(ctx context.Context, key ContextKey, value interface{}) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetValueFromContext will get context with specific key
func GetValueFromContext(ctx context.Context, key ContextKey) interface{} {
	return ctx.Value(key)
}

// ParseTokenClaimFromContext parse token claim from given context
func ParseTokenClaimFromContext(ctx context.Context) *TokenClaim {
	return GetValueFromContext(ctx, ContextKeyTokenClaim).(*TokenClaim)
}

// ParseWorkerKeyFromContext parse token claim from given context
func ParseWorkerKeyFromContext(ctx context.Context) []byte {
	return GetValueFromContext(ctx, ContextKeyWorkerKey).([]byte)
}
