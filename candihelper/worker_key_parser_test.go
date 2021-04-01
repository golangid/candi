package candihelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisPubSubKeyTopic(t *testing.T) {
	got := BuildRedisPubSubKeyTopic("scheduled-notif", map[string]string{"test": "testing"})
	assert.Equal(t, "{\"h\":\"scheduled-notif\",\"message\":\"{\\\"test\\\":\\\"testing\\\"}\"}", got)

	handlerName, message := ParseRedisPubSubKeyTopic(got)
	assert.Equal(t, "scheduled-notif", handlerName)
	assert.Equal(t, "{\"test\":\"testing\"}", message)
}
