package usecase

import (
	"context"

	"github.com/line/line-bot-sdk-go/linebot"
)

// BotUsecase abstraction
type BotUsecase interface {
	ProcessCallback(ctx context.Context, events []*linebot.Event) error
	PushMessageToChannel(ctx context.Context, to, title, message string) error
}
