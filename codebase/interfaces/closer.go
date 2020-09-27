package interfaces

import "context"

// Closer abstraction
type Closer interface {
	Disconnect(ctx context.Context) error
}
