package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Visitor model
type Visitor struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	IPAddress  string             `bson:"ipAddress" json:"ipAddress"`
	UserAgent  string             `bson:"userAgent" json:"userAgent"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	ModifiedAt time.Time          `bson:"modifiedAt" json:"modifiedAt"`
}
