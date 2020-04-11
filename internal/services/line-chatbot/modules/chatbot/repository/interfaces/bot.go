package interfaces

import (
	"context"

	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/domain"
)

// Bot abstraction
type Bot interface {
	ProcessText(ctx context.Context, text string) string
	PushMessageToLine(ctx context.Context, message *domain.Message) error
}
