package repository

import (
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/event/repository/interfaces"
	"github.com/agungdwiprasetyo/backend-microservices/internal/services/wedding/modules/event/repository/mongo"
	mgo "go.mongodb.org/mongo-driver/mongo"
)

// RepoMongo model
type RepoMongo struct {
	EventMongo interfaces.EventRepo
}

// NewRepositoryMongo constructor
func NewRepositoryMongo(readMongo, writeMongo *mgo.Database) *RepoMongo {
	return &RepoMongo{
		EventMongo: mongo.NewEventRepo(readMongo, writeMongo),
	}
}
