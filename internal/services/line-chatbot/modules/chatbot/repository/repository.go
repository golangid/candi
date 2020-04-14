package repository

import (
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/repository/httpcall"
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/repository/interfaces"
	loginterfaces "agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/log/repository/interfaces"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository repo
type Repository struct {
	Translator interfaces.Translator
	Bot        interfaces.Bot
	Event      loginterfaces.Event
	Profile    loginterfaces.Profile
}

// NewRepository constructor
func NewRepository(mongoRead, mongoWrite *mongo.Database) *Repository {
	return &Repository{
		Translator: httpcall.NewTranslatorHTTP(),
		Bot:        httpcall.NewBotHTTP(),
	}
}
