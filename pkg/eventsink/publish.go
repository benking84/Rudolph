// Package eventsink forwards Santa execution events to Pub/Sub.
// A Cloud Storage subscription on that topic replaces the Kinesis Firehose
// archive. Additional subscribers replace the optional Lambda fan-out.
package eventsink

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/pubsub"
)

// Publisher accepts a batch of already-wrapped upload events.
// An empty batch is a success and must not call the remote API.
type Publisher interface {
	Publish(ctx context.Context, machineID string, events []interface{}) error
}

// Nop drops every event. Used when EVENT_TOPIC is unset.
type Nop struct{}

// Publish implements Publisher.
func (Nop) Publish(context.Context, string, []interface{}) error { return nil }

// PubSub publishes one message per event. The ordering key is the machine id,
// which is the same partition choice the Kinesis sink used.
type PubSub struct {
	topic *pubsub.Topic
}

// NewPubSub opens a publisher for topicID in projectID.
func NewPubSub(ctx context.Context, projectID, topicID string) (*PubSub, error) {
	if projectID == "" || topicID == "" {
		return nil, fmt.Errorf("eventsink: project and topic are required")
	}
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("eventsink: pubsub client: %w", err)
	}
	topic := client.Topic(topicID)
	topic.EnableMessageOrdering = true
	return &PubSub{topic: topic}, nil
}

// FromEnv returns a Pub/Sub publisher when EVENT_TOPIC is set, otherwise Nop.
func FromEnv(ctx context.Context) (Publisher, error) {
	topic := os.Getenv("EVENT_TOPIC")
	if topic == "" {
		return Nop{}, nil
	}
	project := os.Getenv("GCP_PROJECT")
	if project == "" {
		project = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	return NewPubSub(ctx, project, topic)
}

// Publish sends each event as JSON. A message attribute machine_id is set so
// subscriptions can filter without parsing the body.
func (p *PubSub) Publish(ctx context.Context, machineID string, events []interface{}) error {
	if len(events) == 0 {
		return nil
	}
	results := make([]*pubsub.PublishResult, 0, len(events))
	for _, event := range events {
		body, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("eventsink: marshal event: %w", err)
		}
		results = append(results, p.topic.Publish(ctx, &pubsub.Message{
			Data:        body,
			OrderingKey: machineID,
			Attributes:  map[string]string{"machine_id": machineID},
		}))
	}
	for _, result := range results {
		if _, err := result.Get(ctx); err != nil {
			p.topic.ResumePublish(machineID)
			return fmt.Errorf("eventsink: publish: %w", err)
		}
	}
	return nil
}
