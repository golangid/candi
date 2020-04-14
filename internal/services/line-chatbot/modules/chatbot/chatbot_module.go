package chatbot

import (
	"os"

	"agungdwiprasetyo.com/backend-microservices/internal/factory/base"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/internal/factory/interfaces"
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/delivery"
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/repository"
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/helper"
	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	// Chatbot service name
	Chatbot constant.Module = iota
)

// Module model
type Module struct {
	restHandler *delivery.RestHandler
}

// NewModule module constructor
func NewModule(params *base.ModuleParam) *Module {

	lineClient, err := linebot.New(os.Getenv("LINE_CHANNEL_SECRET"), os.Getenv("LINE_CHANNEL_TOKEN"))
	if err != nil {
		panic(err)
	}

	repo := repository.NewRepository()
	uc := usecase.NewBotUsecase(lineClient, repo)

	var mod Module
	mod.restHandler = delivery.NewRestHandler(params.Middleware, lineClient, uc)
	return &mod
}

// RestHandler method
func (m *Module) RestHandler(version string) (d interfaces.EchoRestDelivery) {
	switch version {
	case helper.V1:
		d = m.restHandler
	case helper.V2:
		d = nil // TODO versioning
	}
	return
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCDelivery {
	return nil
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return "Chatbot", nil
}

// SubscriberHandler method
func (m *Module) SubscriberHandler(subsType constant.Subscriber) interfaces.SubscriberDelivery {
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Chatbot
}
