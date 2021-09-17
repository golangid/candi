package interfaces

import (
	"context"

	"github.com/golangid/candi/candishared"
)

// Publisher abstract interface
type Publisher interface {
	PublishMessage(ctx context.Context, args *candishared.PublisherArgument) (err error)
}
