package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisPubSubKeyTopic(t *testing.T) {
	got := BuildRedisPubSubKeyTopic("scheduled-notif", "test~123")
	assert.Equal(t, "scheduled-notif~test~123", got)

	prefix, suffix := ParseRedisPubSubKeyTopic(got)
	assert.Equal(t, "scheduled-notif", prefix)
	assert.Equal(t, "test~123", suffix)
}
