package candishared

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEventContext(t *testing.T) {
	event := &EventContext{}

	ctx := context.Background()
	event.SetContext(context.WithValue(ctx, "key1", "value 1"))
	event.SetContext(context.WithValue(ctx, "key2", "value 2"))

	assert.Equal(t, nil, event.Context().Value("key1"))
	assert.Equal(t, "value 2", event.Context().Value("key2"))
}

func TestMultiContext(t *testing.T) {
	event := &EventContext{}
	event.SetContext(context.Background())

	event.SetContextWithValue("key1", "value 1")
	event.SetContextWithValue("key2", "value 2")

	assert.Equal(t, "value 1", event.Context().Value("key1"))
	assert.Equal(t, "value 2", event.Context().Value("key2"))
}
