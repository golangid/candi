package storage

import (
	"context"
)

// Storage abstraction
type Storage interface {
	Upload(ctx context.Context, param *UploadParam) (Response, error)
}
