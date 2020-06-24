package repository

import (
	"os"

	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/firebase"
	"agungdwiprasetyo.com/backend-microservices/internal/notification-service/modules/push-notif/repository/interfaces"
)

// Repository repo
type Repository struct {
	PushNotif interfaces.PushNotif
}

// NewRepository constructor
func NewRepository() *Repository {
	return &Repository{
		PushNotif: firebase.NewFirebaseREST(
			os.Getenv("FIREBASE_HOST"), os.Getenv("FIREBASE_KRAB_KEY"),
		),
	}
}
