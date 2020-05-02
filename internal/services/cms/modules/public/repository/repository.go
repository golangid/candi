package repository

import (
	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/repository/interfaces"
	visitormongo "agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/repository/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

type RepoMongo struct {
	Visitor interfaces.Visitor
}

// NewRepoMongo constructor
func NewRepoMongo(db *mongo.Database) *RepoMongo {
	return &RepoMongo{
		Visitor: visitormongo.NewVisitorRepo(db),
	}
}
