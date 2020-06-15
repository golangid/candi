package usecase

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/storage-service/modules/storage/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

// StorageUsecase abstraction
type StorageUsecase interface {
	Upload(ctx context.Context, buff []byte, metadata *domain.UploadMetadata) <-chan shared.Result
}
