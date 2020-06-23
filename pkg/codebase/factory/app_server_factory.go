package factory

import "context"

// AppServerFactory factory
type AppServerFactory interface {
	Serve()
	Shutdown(ctx context.Context)
}
