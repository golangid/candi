package candihelper

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

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
