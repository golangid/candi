package types

import "context"

// WorkerHandlerFunc types
type WorkerHandlerFunc func(ctx context.Context, message []byte) error

// WorkerErrorHandler types
type WorkerErrorHandler func(ctx context.Context, workerType Worker, workerName string, message []byte, err error)

// WorkerHandlerGroup group of worker handlers by pattern string
type WorkerHandlerGroup struct {
	Handlers []struct {
		Pattern      string
		HandlerFunc  WorkerHandlerFunc
		ErrorHandler []WorkerErrorHandler
	}
}

// Add method from WorkerHandlerGroup, pattern can contains unique topic name, key, and task name
func (m *WorkerHandlerGroup) Add(pattern string, handlerFunc WorkerHandlerFunc, errHandlers ...WorkerErrorHandler) {
	m.Handlers = append(m.Handlers, struct {
		Pattern      string
		HandlerFunc  WorkerHandlerFunc
		ErrorHandler []WorkerErrorHandler
	}{
		Pattern: pattern, HandlerFunc: handlerFunc, ErrorHandler: errHandlers,
	})
}
