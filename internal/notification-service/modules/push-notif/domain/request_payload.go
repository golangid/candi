package domain

// PushNotifRequestPayload model payload
type PushNotifRequestPayload struct {
	To      string `json:"to"`
	Title   string `json:"title"`
	Message string `json:"message"`
}
