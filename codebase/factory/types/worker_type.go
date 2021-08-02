package types

import "context"

type (
	// WorkerHandlerFunc types
	WorkerHandlerFunc func(ctx context.Context, message []byte) error

	// WorkerErrorHandler types
	WorkerErrorHandler func(ctx context.Context, workerType Worker, workerName string, message []byte, err error)

	// WorkerHandler types
	WorkerHandler struct {
		Pattern      string
		HandlerFunc  WorkerHandlerFunc
		ErrorHandler WorkerErrorHandler
		DisableTrace bool
		AutoACK      bool
	}

	// WorkerHandlerOptionFunc types
	WorkerHandlerOptionFunc func(*WorkerHandler)
)

// WorkerHandlerGroup group of worker handlers by pattern string
type WorkerHandlerGroup struct {
	Handlers []WorkerHandler
}

// Add method from WorkerHandlerGroup, pattern can contains unique topic name, key, and task name
func (m *WorkerHandlerGroup) Add(pattern string, handlerFunc WorkerHandlerFunc, opts ...WorkerHandlerOptionFunc) {
	h := WorkerHandler{
		Pattern: pattern, HandlerFunc: handlerFunc, AutoACK: true,
	}

	for _, opt := range opts {
		opt(&h)
	}
	m.Handlers = append(m.Handlers, h)
}

// WorkerHandlerOptionDisableTrace set disable trace
func WorkerHandlerOptionDisableTrace() WorkerHandlerOptionFunc {
	return func(wh *WorkerHandler) {
		wh.DisableTrace = true
	}
}

// WorkerHandlerOptionAutoACK set disable trace
func WorkerHandlerOptionAutoACK(auto bool) WorkerHandlerOptionFunc {
	return func(wh *WorkerHandler) {
		wh.AutoACK = auto
	}
}

// WorkerHandlerOptionAddErrorHandler add error handlers
func WorkerHandlerOptionAddErrorHandler(errHandler WorkerErrorHandler) WorkerHandlerOptionFunc {
	return func(wh *WorkerHandler) {
		wh.ErrorHandler = errHandler
	}
}
