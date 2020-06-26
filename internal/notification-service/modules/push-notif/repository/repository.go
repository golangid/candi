package repository

import (
	"os"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/firebase"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/interfaces"
	redisrepo "agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/redis"
	"github.com/gomodule/redigo/redis"
)

// Repository repo
type Repository struct {
	PushNotif interfaces.PushNotif
	Schedule  interfaces.Schedule
}

// NewRepository constructor
func NewRepository(redisPool *redis.Pool) *Repository {
	return &Repository{
		PushNotif: firebase.NewFirebaseREST(
			os.Getenv("FIREBASE_HOST"), os.Getenv("FIREBASE_KRAB_KEY"),
		),
		Schedule: redisrepo.NewRedisRepo(redisPool),
	}
}
