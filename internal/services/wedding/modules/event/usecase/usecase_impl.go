package usecase

import (
	"context"

	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/event/domain"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/event/repository"
)

type eventUsecaseImpl struct {
	repo *repository.RepoMongo
}

// NewEventUsecase create new customer usecase
func NewEventUsecase(repo *repository.RepoMongo) EventUsecase {
	return &eventUsecaseImpl{
		repo: repo,
	}
}

func (uc *eventUsecaseImpl) FindByCode(ctx context.Context, code string) (*domain.Event, error) {
	repoRes := <-uc.repo.EventMongo.Find(ctx, domain.Event{Code: code})
	if repoRes.Error != nil {
		return nil, repoRes.Error
	}

	return repoRes.Data.(*domain.Event), nil
}
