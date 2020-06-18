package chatbot

import (
	"os"

	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/chatbot/delivery"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/chatbot/repository"
	"agungdwiprasetyo.com/backend-microservices/internal/line-chatbot/modules/chatbot/usecase"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/base"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/factory/constant"
	"agungdwiprasetyo.com/backend-microservices/pkg/codebase/interfaces"
	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	// Chatbot service name
	Chatbot constant.Module = "Chatbot"
)

// Module model
type Module struct {
	restHandler *delivery.RestHandler
}

// NewModule module constructor
func NewModule(deps *base.Dependency) *Module {

	lineClient, err := linebot.New(os.Getenv("LINE_CHANNEL_SECRET"), os.Getenv("LINE_CHANNEL_TOKEN"))
	if err != nil {
		panic(err)
	}

	repo := repository.NewRepository(deps.Config.MongoRead, deps.Config.MongoWrite)
	uc := usecase.NewBotUsecase(lineClient, repo)

	var mod Module
	mod.restHandler = delivery.NewRestHandler(deps.Middleware, uc)
	return &mod
}

// RestHandler method
func (m *Module) RestHandler() interfaces.EchoRestHandler {
	return m.restHandler
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return nil
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() (name string, resolver interface{}) {
	return string(Chatbot), nil
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType constant.Worker) interfaces.WorkerHandler {
	return nil
}

// Name get module name
func (m *Module) Name() constant.Module {
	return Chatbot
}
