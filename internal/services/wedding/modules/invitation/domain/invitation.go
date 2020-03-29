package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Invitation model
type Invitation struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	WaNumber  string             `json:"waNumber" bson:"wa_number"`
	Message   string             `json:"message" bson:"message"`
	Relation  string             `json:"relation" bson:"relation"`
	IsAttend  bool               `json:"isAttend" bson:"is_attend"`
	CreatedAt time.Time          `json:"created" bson:"created"`
}

// Event model
type Event struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Code      string             `json:"code" bson:"code"`
	Date      string             `bson:"date" json:"date"`
	CountDown int                `bson:"-" json:"countDown"`
	Ceremony  string             `bson:"ceremony" json:"ceremony"`
	Reception string             `bson:"reception" json:"reception"`
	Address   string             `bson:"address" json:"address"`
}
