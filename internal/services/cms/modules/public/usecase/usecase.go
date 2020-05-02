package usecase

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

type PublicUsecase interface {
	GetHomePage(ctx context.Context) *domain.HomePage
	GetAllVisitor(ctx context.Context, filter *shared.Filter) (data []domain.Visitor, meta *shared.Meta, err error)
}
