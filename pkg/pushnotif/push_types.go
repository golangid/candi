package pushnotif

// PushRequest model
type PushRequest struct {
	To           string                 `json:"to"`
	Notification *Notification          `json:"notification"`
	Data         map[string]interface{} `json:"data"`
}

// Notification model
type Notification struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	Image          string `json:"image"`
	Sound          string `json:"sound"`
	MutableContent bool   `json:"mutable-content"`
	ResourceID     string `json:"resourceId"`
	ResourceName   string `json:"resoureceName"`
}

// PushResponse response data
type PushResponse struct {
	Error bool        `json:"error"`
	Body  interface{} `json:"body"`
}
