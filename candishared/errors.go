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
	Delay          time.Duration
	Retry          int
	Message        string
	NewArgsPayload []byte
}

// Error implement error
func (e *ErrorRetrier) Error() string {
	return e.Message
}
