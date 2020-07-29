package mongo

import (
	"context"
	"errors"
	"time"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/domain"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"agungdwiprasetyo.com/backend-microservices/pkg/logger"
	"agungdwiprasetyo.com/backend-microservices/pkg/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SubscriberRepoMongo mongo repo
type SubscriberRepoMongo struct {
	read, write *mongo.Database
	collection  string
}

// NewSubscriberRepoMongo constructor
func NewSubscriberRepoMongo(read, write *mongo.Database) *SubscriberRepoMongo {
	return &SubscriberRepoMongo{
		read: read, write: write, collection: "topics",
	}
}

func (r *SubscriberRepoMongo) Save(ctx context.Context, data *domain.Topic) <-chan error {
	output := make(chan error)

	go func() {
		defer close(output)

		var err error

		data.ModifiedAt = time.Now()
		if data.ID == "" {
			data.ID = primitive.NewObjectID().Hex()
			data.CreatedAt = time.Now()
			_, err = r.write.Collection(r.collection).InsertOne(ctx, data)
		} else {
			opt := options.UpdateOptions{
				Upsert: helper.ToBoolPtr(true),
			}
			_, err = r.write.Collection(r.collection).UpdateOne(ctx,
				bson.M{
					"_id": data.ID,
				},
				bson.M{
					"$set": data,
				}, &opt)
		}

		if err != nil {
			output <- err
			return
		}
	}()

	return output
}

func (r *SubscriberRepoMongo) FindTopic(ctx context.Context, where domain.Topic) <-chan shared.Result {
	output := make(chan shared.Result)

	go func() {
		defer close(output)
		var data domain.Topic

		bsonWhere := make(primitive.M)
		if where.ID != "" {
			bsonWhere["_id"], _ = primitive.ObjectIDFromHex(where.ID)
		}
		if where.Name != "" {
			bsonWhere["name"] = where.Name
		}

		err := r.read.Collection(r.collection).FindOne(ctx, bsonWhere).Decode(&data)
		if err != nil {
			output <- shared.Result{Error: err}
			return
		}

		output <- shared.Result{Data: data}
	}()

	return output
}

func (r *SubscriberRepoMongo) RemoveSubscriber(ctx context.Context, subscriber *domain.Subscriber) <-chan error {
	output := make(chan error)

	go func() {
		defer close(output)

		bsonWhere := make(primitive.M)
		bsonWhere["name"] = subscriber.Topic
		bsonWhere["subscribers.id"] = subscriber.ID

		_, err := r.write.Collection(r.collection).UpdateOne(ctx,
			bsonWhere,
			primitive.M{
				"$set": primitive.M{"subscribers.$.isActive": false},
			}, &options.UpdateOptions{
				Upsert: helper.ToBoolPtr(true),
			})
		if err != nil {
			logger.LogE(err.Error())
			output <- err
			return
		}

	}()

	return output
}

func (r *SubscriberRepoMongo) FindSubscriber(ctx context.Context, topicName string, subscriber *domain.Subscriber) <-chan shared.Result {
	output := make(chan shared.Result)

	go func() {
		defer close(output)
		var data domain.Topic

		bsonWhere := make(primitive.M)
		bsonWhere["name"] = topicName
		bsonWhere["subscribers.id"] = subscriber.ID

		err := r.read.Collection(r.collection).FindOne(ctx, bsonWhere).Decode(&data)
		if err != nil {
			logger.LogE(err.Error())
			output <- shared.Result{Error: err}
			return
		}

		if len(data.Subscribers) == 0 {
			output <- shared.Result{Error: errors.New("Not found")}
			return
		}

		output <- shared.Result{Data: data.Subscribers[0]}
	}()

	return output
}
