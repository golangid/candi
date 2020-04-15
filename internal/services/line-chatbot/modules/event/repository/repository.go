package repository

import (
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/event/repository/interfaces"
	eventmongo "agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/event/repository/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

// RepoMongo mongo
type RepoMongo struct {
	Event interfaces.Event
}

// NewRepoMongo constructor
func NewRepoMongo(db *mongo.Database) *RepoMongo {
	return &RepoMongo{
		Event: eventmongo.NewEventRepo(db),
	}
}
