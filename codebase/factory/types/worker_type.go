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
		Configs      map[string]interface{}
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

// WorkerHandlerOptionAutoACK set auto ACK
func WorkerHandlerOptionAutoACK(auto bool) WorkerHandlerOptionFunc {
	return func(wh *WorkerHandler) {
		wh.AutoACK = auto
	}
}

// WorkerHandlerOptionAddConfig set config
func WorkerHandlerOptionAddConfig(key string, value interface{}) WorkerHandlerOptionFunc {
	return func(wh *WorkerHandler) {
		if wh.Configs == nil {
			wh.Configs = map[string]interface{}{}
		}
		wh.Configs[key] = value
	}
}

// WorkerHandlerOptionAddHandlers add after handlers execute after main handler
func WorkerHandlerOptionAddHandlers(handlerFuncs ...WorkerHandlerFunc) WorkerHandlerOptionFunc {
	return func(wh *WorkerHandler) {
		wh.HandlerFuncs = append(wh.HandlerFuncs, handlerFuncs...)
	}
}
