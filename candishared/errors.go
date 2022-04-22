package candishared

import (
	"time"
)

// GraphQLErrorResolver graphql error with extensions
type GraphQLErrorResolver interface {
	Error() string
	Extensions() map[string]interface{}
}

type resolveErrorImpl struct {
	message    string
	extensions map[string]interface{}
}

func (r *resolveErrorImpl) Error() string {
	return r.message
}
func (r *resolveErrorImpl) Extensions() map[string]interface{} {
	return r.extensions
}

// NewGraphQLErrorResolver constructor
func NewGraphQLErrorResolver(errMesage string, extensions map[string]interface{}) GraphQLErrorResolver {
	return &resolveErrorImpl{
		message: errMesage, extensions: extensions,
	}
}

// ErrorRetrier task queue worker for retry error with retry count and delay between retry
type ErrorRetrier struct {
	// Delay run retry, skip retry if delay <= 0
	Delay    time.Duration
	NewRetry int

	// Message for error value
	Message string

	// NewArgsPayload overide args message payload
	NewArgsPayload []byte

	StackTrace string

	NewRetryIntervalFunc func(retries int) time.Duration
}

// Error implement error
func (e *ErrorRetrier) Error() string {
	return e.Message
}
