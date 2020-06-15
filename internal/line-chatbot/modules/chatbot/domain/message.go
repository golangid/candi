package domain

// Message domain model
type Message struct {
	To       string           `json:"to"`
	Messages []ContentMessage `json:"messages"`
}

type ContentMessage struct {
	Type     string        `json:"type"`
	AltText  string        `json:"altText"`
	Contents ContentFormat `json:"contents"`
}

type ContentFormat struct {
	Type string      `json:"type"`
	Body ContentBody `json:"body"`
}

type ContentBody struct {
	Type     string    `json:"type"`
	Layout   string    `json:"layout"`
	Contents []Content `json:"contents"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
