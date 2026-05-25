package mailer

import "context"

type Sender interface {
	Send(ctx context.Context, email, subject, body string) error
}

type NoopSender struct{}

func (NoopSender) Send(context.Context, string, string, string) error {
	return nil
}
