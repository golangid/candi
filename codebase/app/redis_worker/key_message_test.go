package redisworker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisPubSubMessage(t *testing.T) {
	got := CreateRedisPubSubMessage("scheduled-notif", map[string]string{"test": "testing"})
	assert.Equal(t, "{\"h\":\"scheduled-notif\",\"message\":\"{\\\"test\\\":\\\"testing\\\"}\"}", got)

	msg := ParseRedisPubSubKeyTopic([]byte(got))
	assert.Equal(t, "scheduled-notif", msg.HandlerName)
	assert.Equal(t, "{\"test\":\"testing\"}", msg.Message)
}
