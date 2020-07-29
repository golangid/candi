package domain

import "time"

type Topic struct {
	ID          string        `bson:"_id" json:"id"`
	Name        string        `bson:"name" json:"name"`
	Subscribers []*Subscriber `bson:"subscribers" json:"subscribers"`
	CreatedAt   time.Time     `bson:"createdAt" json:"createdAt"`
	ModifiedAt  time.Time     `bson:"modifiedAt" json:"modifiedAt"`
}

type Subscriber struct {
	ID         string    `bson:"id" json:"id"`
	Topic      string    `bson:"-" json:"topic"`
	Name       string    `bson:"name" json:"name"`
	IsActive   bool      `bson:"isActive" json:"isActive"`
	CreatedAt  time.Time `bson:"createdAt" json:"createdAt"`
	ModifiedAt time.Time `bson:"modifiedAt" json:"modifiedAt"`

	Stop   <-chan struct{} `bson:"-" json:"-"`
	Events chan<- *Event   `bson:"-" json:"-"`
}
