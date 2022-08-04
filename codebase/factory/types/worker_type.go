package types

import (
	"github.com/golangid/candi/candishared"
)

type (
	// WorkerHandlerFunc types
	WorkerHandlerFunc func(ctx *candishared.EventContext) error

	// WorkerHandler types
	WorkerHandler struct {
		Pattern      string
		HandlerFuncs []WorkerHandlerFunc
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

// Add method from WorkerHandlerGroup, patternRoute can contains unique topic name, key, or task name
func (m *WorkerHandlerGroup) Add(patternRoute string, mainHandlerFunc WorkerHandlerFunc, opts ...WorkerHandlerOptionFunc) {
	h := WorkerHandler{
		Pattern: patternRoute, HandlerFuncs: []WorkerHandlerFunc{mainHandlerFunc}, AutoACK: true,
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

// WorkerHandlerOptionAddHandlers add after handlers execute after main handler
func WorkerHandlerOptionAddHandlers(handlerFuncs ...WorkerHandlerFunc) WorkerHandlerOptionFunc {
	return func(wh *WorkerHandler) {
		wh.HandlerFuncs = append(wh.HandlerFuncs, handlerFuncs...)
	}
}
