package usecase

import (
	"context"
	"net/http"

	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/repository"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

type publicUsecaseImpl struct {
	repo *repository.RepoMongo
}

// NewPublicUsecase constructor
func NewPublicUsecase(repo *repository.RepoMongo) PublicUsecase {
	return &publicUsecaseImpl{
		repo: repo,
	}
}

func (uc *publicUsecaseImpl) GetHomePage(ctx context.Context) *domain.HomePage {
	var visitor domain.Visitor

	headers := ctx.Value(shared.ContextKey("headers")).(http.Header)
	visitor.IPAddress = headers.Get("X-Real-IP")
	visitor.UserAgent = headers.Get("User-Agent")

	<-uc.repo.Visitor.Save(ctx, &visitor)

	return &domain.HomePage{
		Content: `Hello, my name is Agung Dwi Prasetyo
		This page is still under development
		See my GitHub resume
		Or my Curriculum Vitae`,
		Skills: []string{"Golang", "GRPC", "GraphQL", "Kafka"},
		Footer: "test",
	}
}
