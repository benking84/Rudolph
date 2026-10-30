package eventsink

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNopPublish(t *testing.T) {
	err := Nop{}.Publish(context.Background(), "machine", []interface{}{"event"})
	assert.NoError(t, err)
}

func TestPubSubPublishEmptyIsNoop(t *testing.T) {
	// topic is nil. An empty batch must return before it is used.
	p := &PubSub{}
	assert.NoError(t, p.Publish(context.Background(), "machine", nil))
	assert.NoError(t, p.Publish(context.Background(), "machine", []interface{}{}))
}
