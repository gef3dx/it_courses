package queue

import "context"

type Message struct {
	Type     string         `json:"type"`
	Payload  map[string]any `json:"payload"`
	Metadata map[string]any `json:"metadata"`
}

type Publisher interface {
	Publish(ctx context.Context, message Message) error
}

type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, Message) error {
	return nil
}
