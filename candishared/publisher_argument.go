package candishared

import (
	"errors"
	"time"
)

// PublisherArgument declare publisher argument
type PublisherArgument struct {
	// Topic or queue name
	Topic       string
	Key         string
	Header      map[string]interface{}
	ContentType string
	Message     []byte
	Delay       time.Duration

	// Deprecated : use Message
	Data interface{}
}

func (p *PublisherArgument) Validate() error {
	if len(p.Topic) == 0 {
		return errors.New("topic cannot empty")
	}
	if len(p.Message) == 0 {
		return errors.New("message cannot empty")
	}

	return nil
}
