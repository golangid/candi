package interfaces

// Broker abstraction
type Broker interface {
	GetConfiguration() interface{} // get broker configuration (different type for each broker)
	GetPublisher() Publisher
	Health() map[string]error
	Closer
}
