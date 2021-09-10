package interfaces

import (
	"context"
	"time"
)

// Cache abstract interface
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	GetKeys(ctx context.Context, pattern string) ([]string, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Set(ctx context.Context, key string, value interface{}, expire time.Duration) error
	Exists(ctx context.Context, key string) (bool, error)
	Delete(ctx context.Context, key string) error
}
