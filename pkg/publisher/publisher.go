package publisher

import "context"

// Publisher abstract interface
type Publisher interface {
	PublishMessage(ctx context.Context, topic, key string, message interface{}) (err error)
}
