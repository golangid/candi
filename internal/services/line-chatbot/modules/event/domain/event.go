package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event domain model
type Event struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	ReplyToken string             `bson:"reply_token" json:"replyToken"`
	Type       string             `bson:"type" json:"type"`
	Timestamp  time.Time          `bson:"timestamp" json:"timestamp"`
	SourceID   string             `bson:"source_id" json:"sourceId"`
	SourceType string             `bson:"source_type" json:"sourceType"`
	Message    struct {
		ID       string `bson:"id" json:"id"`
		Type     string `bson:"type" json:"type"`
		Text     string `bson:"text" json:"text"`
		Response string `bson:"response" json:"response"`
	} `bson:"message" json:"message"`
	Error *string `bson:"error" json:"error"`
}
