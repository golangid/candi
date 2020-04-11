package domain

import (
	"time"
)

// Profile model
type Profile struct {
	ID            string    `bson:"id" json:"id"`
	Type          string    `bson:"type" json:"type"`
	Name          string    `bson:"name" json:"name"`
	Avatar        string    `bson:"avatar" json:"avatar"`
	StatusMessage string    `bson:"status_message" json:"statusMessage"`
	CreatedAt     time.Time `bson:"createdAt" json:"createdAt"`
	ModifiedAt    time.Time `bson:"modifiedAt" json:"modifiedAt"`
}
