package mongo

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/wedding/modules/event/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// EventMongoRepo mongo repository
type EventMongoRepo struct {
	readDB, writeDB *mongo.Database
	collection      string
}

// NewEventRepo create new event repository
func NewEventRepo(readDB, writeDB *mongo.Database) *EventMongoRepo {
	return &EventMongoRepo{
		readDB:     readDB,
		writeDB:    writeDB,
		collection: "events",
	}
}

// Find data customer
func (r *EventMongoRepo) Find(ctx context.Context, where domain.Event) <-chan shared.Result {
	output := make(chan shared.Result)

	go func() {
		defer close(output)

		var result shared.Result
		coll := r.readDB.Collection(r.collection)
		var event domain.Event
		err := coll.FindOne(ctx, bson.M{
			"code": where.Code,
		}).Decode(&event)
		if err != nil {
			result.Error = err
		}
		result.Data = &event

		output <- result
	}()

	return output
}
