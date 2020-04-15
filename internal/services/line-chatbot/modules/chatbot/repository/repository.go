package repository

import (
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/repository/httpcall"
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/repository/interfaces"
	eventinterfaces "agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/event/repository/interfaces"
	eventmongo "agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/event/repository/mongo"
	"go.mongodb.org/mongo-driver/mongo"
)

// Repository repo
type Repository struct {
	Translator interfaces.Translator
	Bot        interfaces.Bot
	Event      eventinterfaces.Event
	Profile    eventinterfaces.Profile
}

// NewRepository constructor
func NewRepository(mongoRead, mongoWrite *mongo.Database) *Repository {
	return &Repository{
		Translator: httpcall.NewTranslatorHTTP(),
		Bot:        httpcall.NewBotHTTP(),
		Event:      eventmongo.NewEventRepo(mongoWrite),
	}
}
