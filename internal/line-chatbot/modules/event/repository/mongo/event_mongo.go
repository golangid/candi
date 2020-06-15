package mongo

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/event/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EventMongoRepo repo
type EventMongoRepo struct {
	db         *mongo.Database
	collection string
}

// NewEventRepo create new customer repository
func NewEventRepo(db *mongo.Database) *EventMongoRepo {
	return &EventMongoRepo{
		db:         db,
		collection: "events",
	}
}

func (r *EventMongoRepo) FindAll(ctx context.Context, filter *shared.Filter) <-chan shared.Result {
	output := make(chan shared.Result)
	go func() {
		defer close(output)

		findOptions := options.Find()

		var sort = 1
		if filter.Sort == "desc" {
			sort = -1
		}
		if filter.OrderBy != "" {
			findOptions.SetSort(map[string]int{filter.OrderBy: sort})
		}
		findOptions.SetLimit(int64(filter.Limit))
		findOptions.SetSkip(int64(filter.Offset))

		cursor, err := r.db.Collection(r.collection).Find(ctx, bson.M{}, findOptions)
		if err != nil {
			output <- shared.Result{Error: err}
			return
		}

		var results []domain.Event
		for cursor.Next(ctx) {
			var res domain.Event
			if err := cursor.Decode(&res); err != nil {
				output <- shared.Result{Error: err}
				return
			}
			results = append(results, res)
		}

		output <- shared.Result{Data: results}
	}()
	return output
}

func (r *EventMongoRepo) Count(ctx context.Context, filter *shared.Filter) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)

		count, _ := r.db.Collection(r.collection).CountDocuments(ctx, bson.M{})
		output <- int(count)
	}()
	return output
}

func (r *EventMongoRepo) Save(ctx context.Context, data *domain.Event) <-chan error {
	output := make(chan error)
	go func() {
		defer close(output)
		var err error

		if data.ID.IsZero() {
			data.ID = primitive.NewObjectID()
			_, err = r.db.Collection(r.collection).InsertOne(ctx, data)
		} else {
			opt := options.UpdateOptions{
				Upsert: helper.ToBoolPtr(true),
			}
			_, err = r.db.Collection(r.collection).UpdateOne(ctx,
				bson.M{
					"_id": data.ID,
				},
				bson.M{
					"$set": data,
				}, &opt)
		}

		output <- err
	}()
	return output
}
