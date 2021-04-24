package interfaces

import (
	"context"

	"pkg.agungdp.dev/candi/candishared"
)

// Publisher abstract interface
type Publisher interface {
	PublishMessage(ctx context.Context, args *candishared.PublisherArgument) (err error)
}
