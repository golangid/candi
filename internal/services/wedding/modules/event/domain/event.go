package domain

// Event model
type Event struct {
	ID        string `json:"id" bson:"_id"`
	Code      string `json:"code" bson:"code"`
	Date      string `bson:"date" json:"date"`
	CountDown int    `bson:"countDown" json:"countDown"`
	Ceremony  string `bson:"ceremony" json:"ceremony"`
	Reception string `bson:"reception" json:"reception"`
	Address   string `bson:"address" json:"address"`
}
