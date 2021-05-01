package factory

import "context"

// AppServerFactory factory for server and/or worker abstraction
type AppServerFactory interface {
	Serve()
	Shutdown(ctx context.Context)
	Name() string
}
