package usecase

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/domain"
)

type PublicUsecase interface {
	GetHomePage(ctx context.Context) *domain.HomePage
}
