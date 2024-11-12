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
		Configs      map[string]any
	}

	// WorkerHandlerOptionFunc types
	WorkerHandlerOptionFunc func(*WorkerHandler)
)

// WorkerHandlerGroup group of worker handlers by pattern string
type WorkerHandlerGroup struct {
	Handlers []WorkerHandler
}

// AddMultiRoute method from WorkerHandlerGroup to handle multi topic in single method
func (m *WorkerHandlerGroup) AddMultiRoute(patternRoutes []string, mainHandlerFunc WorkerHandlerFunc, opts ...WorkerHandlerOptionFunc) {
	for _, route := range patternRoutes {
		if len(route) == 0 {
			continue
		}
		m.Add(route, mainHandlerFunc, opts...)
	}
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
func WorkerHandlerOptionAddConfig(key string, value any) WorkerHandlerOptionFunc {
	return func(wh *WorkerHandler) {
		if wh.Configs == nil {
			wh.Configs = map[string]any{}
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
