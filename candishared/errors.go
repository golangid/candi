package candishared

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

type (
	// GraphQLErrorResolver graphql error with extensions
	GraphQLErrorResolver interface {
		Error() string
		Extensions() map[string]any
	}

	// MultiError abstract interface
	MultiError interface {
		Append(key string, err error) MultiError
		HasError() bool
		IsNil() bool
		Clear()
		ToMap() map[string]string
		Merge(MultiError) MultiError
		Error() string
		ToGraphQLExtension(mainErr string) GraphQLErrorResolver
	}
)

type resolveErrorImpl struct {
	message    string
	extensions map[string]any
}

func (r *resolveErrorImpl) Error() string {
	return r.message
}
func (r *resolveErrorImpl) Extensions() map[string]any {
	return r.extensions
}

// NewGraphQLErrorResolver constructor
func NewGraphQLErrorResolver(errMesage string, extensions map[string]any) GraphQLErrorResolver {
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

type (
	multiError struct {
		lock sync.Mutex
		errs map[string]string
	}
)

// NewMultiError constructor
func NewMultiError() MultiError {
	return &multiError{errs: make(map[string]string)}
}

// Append error to multierror
func (m *multiError) Append(key string, err error) MultiError {
	m.lock.Lock()
	defer m.lock.Unlock()
	if err != nil {
		m.errs[key] = err.Error()
	}
	return m
}

// HasError check if err is exist
func (m *multiError) HasError() bool {
	return len(m.errs) != 0
}

// IsNil check if err is nil
func (m *multiError) IsNil() bool {
	return len(m.errs) == 0
}

// Clear make empty list of errors
func (m *multiError) Clear() {
	m.errs = map[string]string{}
}

// ToMap return list map of error
func (m *multiError) ToMap() map[string]string {
	return m.errs
}

// Merge from another multi error
func (m *multiError) Merge(e MultiError) MultiError {
	for k, v := range e.ToMap() {
		m.Append(k, errors.New(v))
	}
	return m
}

// Error implement error from multiError
func (m *multiError) Error() string {
	var str []string
	for i, s := range m.errs {
		str = append(str, fmt.Sprintf("%s: %s", i, s))
	}
	return strings.Join(str, "\n")
}

// ToGraphQLExtension transform to graphql error extension
func (m *multiError) ToGraphQLExtension(mainErr string) GraphQLErrorResolver {
	gqlErr := &resolveErrorImpl{
		message:    mainErr,
		extensions: make(map[string]any),
	}
	for k, v := range m.errs {
		gqlErr.extensions[k] = v
	}
	return gqlErr
}
