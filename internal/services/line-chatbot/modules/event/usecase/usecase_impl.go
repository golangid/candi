package usecase

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/event/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/event/repository"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
)

type eventUsecaseImpl struct {
	repo *repository.RepoMongo
}

// NewEventUsecase constructor
func NewEventUsecase(repo *repository.RepoMongo) EventUsecase {
	return &eventUsecaseImpl{
		repo: repo,
	}
}

func (uc *eventUsecaseImpl) FindAll(ctx context.Context, filter *shared.Filter) (events []domain.Event, meta *shared.Meta, err error) {
	filter.CalculateOffset()

	count := uc.repo.Event.Count(ctx, filter)
	repoRes := <-uc.repo.Event.FindAll(ctx, filter)
	if repoRes.Error != nil {
		err = repoRes.Error
		return
	}

	events = repoRes.Data.([]domain.Event)
	meta = shared.NewMeta(int(filter.Page), int(filter.Limit), <-count)

	return
}
