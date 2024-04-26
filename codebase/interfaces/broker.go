package interfaces

import "github.com/golangid/candi/codebase/factory/types"

// Broker abstraction
type Broker interface {
	GetPublisher() Publisher
	GetName() types.Worker
	Health() map[string]error
	Closer
}
