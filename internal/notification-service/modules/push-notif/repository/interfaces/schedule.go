package interfaces

import (
	"context"
	"time"
)

// Schedule abstraction
type Schedule interface {
	SaveScheduledNotification(ctx context.Context, key string, data []byte, duration time.Duration) error
}
