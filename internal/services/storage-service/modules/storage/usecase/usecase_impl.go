package usecase

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"agungdwiprasetyo.com/backend-microservices/internal/services/storage-service/modules/storage/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"github.com/minio/minio-go/v6"
)

type storageUsecaseImpl struct {
	minioClient *minio.Client
}

// NewStorageUsecase constructor
func NewStorageUsecase() StorageUsecase {
	minioClient, err := minio.New(os.Getenv("MINIO_HOST"), os.Getenv("MINIO_ACCESS_KEY"), os.Getenv("MINIO_SECRET_KEY"), false)
	if err != nil {
		panic(err)
	}
	return &storageUsecaseImpl{
		minioClient: minioClient,
	}
}

func (uc *storageUsecaseImpl) Upload(ctx context.Context, buff []byte, metadata *domain.UploadMetadata) <-chan shared.Result {
	output := make(chan shared.Result)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				output <- shared.Result{Error: fmt.Errorf("%v", r)}
			}
			close(output)
		}()

		n, err := uc.minioClient.PutObject("tong", metadata.Filename, bytes.NewReader(buff), -1,
			minio.PutObjectOptions{ContentType: metadata.ContentType})
		if err != nil {
			logger.LogE(err.Error())
			panic(err)
		}

		fmt.Println("Uploaded", " size: ", n, "Successfully.", "localhost:9000/tong/...")
	}()

	return output
}
