package mongo

import (
	"context"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/services/cms/modules/public/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// VisitorMongoRepo repo
type VisitorMongoRepo struct {
	db         *mongo.Database
	collection string
}

// NewVisitorRepo create new customer repository
func NewVisitorRepo(db *mongo.Database) *VisitorMongoRepo {
	return &VisitorMongoRepo{
		db:         db,
		collection: "visitors",
	}
}

func (r *VisitorMongoRepo) FindAll(ctx context.Context, filter *shared.Filter) <-chan shared.Result {
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

		var results []domain.Visitor
		for cursor.Next(ctx) {
			var res domain.Visitor
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

func (r *VisitorMongoRepo) Count(ctx context.Context, filter *shared.Filter) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)

		count, _ := r.db.Collection(r.collection).CountDocuments(ctx, bson.M{})
		output <- int(count)
	}()
	return output
}

func (r *VisitorMongoRepo) Save(ctx context.Context, data *domain.Visitor) <-chan error {
	output := make(chan error)
	go func() {
		defer close(output)
		var err error

		data.ModifiedAt = time.Now()
		if data.ID.IsZero() {
			data.ID = primitive.NewObjectID()
			data.CreatedAt = time.Now()
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
