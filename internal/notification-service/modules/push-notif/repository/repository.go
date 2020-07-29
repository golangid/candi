package repository

import (
	"os"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/firebase"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/interfaces"
	smongo "agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/mongo"
	redisrepo "agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/redis"
	"github.com/gomodule/redigo/redis"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository repo
type Repository struct {
	PushNotif  interfaces.PushNotif
	Schedule   interfaces.Schedule
	Subscriber interfaces.Subscriber
}

// NewRepository constructor
func NewRepository(redisPool *redis.Pool, dbReadMongo, dbWriteMongo *mongo.Database) *Repository {
	return &Repository{
		PushNotif: firebase.NewFirebaseREST(
			os.Getenv("FIREBASE_HOST"), os.Getenv("FIREBASE_KRAB_KEY"),
		),
		Schedule:   redisrepo.NewRedisRepo(redisPool),
		Subscriber: smongo.NewSubscriberRepoMongo(dbReadMongo, dbWriteMongo),
	}
}
