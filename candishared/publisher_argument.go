package candishared

// PublisherArgument declare publisher argument
type PublisherArgument struct {
	// Topic or queue name
	Topic       string
	Key         string
	Header      map[string]interface{}
	ContentType string
	// Deprecated : use Message
	Data    interface{}
	Message []byte
}
