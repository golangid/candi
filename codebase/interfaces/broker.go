package interfaces

import "pkg.agungdp.dev/candi/codebase/factory/types"

// Broker abstraction
type Broker interface {
	GetConfiguration() interface{} // get broker configuration (different type for each broker)
	GetPublisher() Publisher
	GetName() types.Worker
	Health() map[string]error
	Closer
}
