package queue

import "context"

type FallbackQueue struct{}

func (FallbackQueue) Publish(context.Context, Message) error {
	return nil
}
