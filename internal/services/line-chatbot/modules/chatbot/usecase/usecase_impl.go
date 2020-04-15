package usecase

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"

	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/chatbot/repository"
	eventdomain "agungdwiprasetyo.com/backend-microservices/internal/services/line-chatbot/modules/event/domain"
	"github.com/line/line-bot-sdk-go/linebot"
)

type botUsecaseImpl struct {
	lineClient *linebot.Client
	repo       *repository.Repository
}

// NewBotUsecase constructor
func NewBotUsecase(client *linebot.Client, repo *repository.Repository) BotUsecase {
	uc := new(botUsecaseImpl)
	uc.lineClient = client
	uc.repo = repo
	return uc
}

func (uc *botUsecaseImpl) ProcessCallback(ctx context.Context, events []*linebot.Event) error {
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			fmt.Println(r)
		}
	}()

	for _, event := range events {
		var (
			profile  eventdomain.Profile
			eventLog eventdomain.Event
		)

		profileResp, err := uc.lineClient.GetProfile(event.Source.UserID).Do()
		if err == nil {
			profile.Type = string(event.Source.Type)
			switch profile.Type {
			case "user":
				profile.ID = event.Source.UserID
			case "group":
				profile.ID = event.Source.GroupID
			}
			profile.Name = profileResp.DisplayName
			profile.Avatar = profileResp.PictureURL
			profile.StatusMessage = profileResp.StatusMessage
		}

		// save event
		eventLog.ReplyToken = event.ReplyToken
		eventLog.Type = string(event.Type)
		eventLog.Timestamp = event.Timestamp
		eventLog.SourceID = profile.ID
		eventLog.SourceType = profile.Type

		switch event.Type {
		case linebot.EventTypeJoin:
			uc.ReplyMessage(event, fmt.Sprintf("Hello %s :)", profileResp.DisplayName))

		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:

				var responseText string
				var inputText = message.Text
				translateToEnglish, translateToIndonesian := "terjemahkan ini ke inggris:", "terjemahkan ini ke indonesia:"

				switch {
				case strings.HasPrefix(strings.ToLower(inputText), translateToEnglish):
					i := strings.Index(strings.ToLower(inputText), translateToEnglish)
					if i >= 0 {
						inputText = inputText[i+len(translateToEnglish):]
					}
					responseText = uc.repo.Translator.Translate(ctx, "id", "en", inputText)

				case strings.HasPrefix(strings.ToLower(inputText), translateToIndonesian):
					i := strings.Index(strings.ToLower(inputText), translateToIndonesian)
					if i >= 0 {
						inputText = inputText[i+len(translateToIndonesian):]
					}
					responseText = uc.repo.Translator.Translate(ctx, "en", "id", inputText)

				default:
					responseText = uc.repo.Bot.ProcessText(ctx, inputText)
				}

				responseText = strings.TrimSpace(responseText)
				err := uc.ReplyMessage(event, responseText)

				eventLog.Message.ID = message.ID
				eventLog.Message.Text = message.Text
				eventLog.Message.Response = responseText
				eventLog.Error = err
			}
		}

		<-uc.repo.Event.Save(ctx, &eventLog)
		// uc.repo.Profile.Save(ctx, &profile)
	}

	return nil
}

func (uc *botUsecaseImpl) ReplyMessage(event *linebot.Event, messages ...string) error {
	var lineMessages []linebot.SendingMessage
	for _, msg := range messages {
		lineMessages = append(lineMessages, linebot.NewTextMessage(msg))
	}

	if _, err := uc.lineClient.ReplyMessage(event.ReplyToken, lineMessages...).Do(); err != nil {
		return err
	}

	return nil
}

func (uc *botUsecaseImpl) PushMessageToChannel(ctx context.Context, to, title, message string) error {
	var lineMessage domain.Message

	lineMessage.To = to
	lineMessage.Messages = append(lineMessage.Messages, domain.ContentMessage{
		Type: "flex", AltText: title, Contents: domain.ContentFormat{
			Type: "bubble", Body: domain.ContentBody{
				Type: "box", Layout: "horizontal", Contents: []domain.Content{
					domain.Content{
						Type: "text", Text: message,
					},
				},
			},
		},
	})

	return uc.repo.Bot.PushMessageToLine(ctx, &lineMessage)
}
