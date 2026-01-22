package email

import (
	"context"
)

type Sender interface {
	SendWelcomeEmail(ctx context.Context, email string, name string) error
	SendPasswordResetEmail(ctx context.Context, email, baseURL, token string) error
}

var (
	_ Sender = (*NullSender)(nil)
	_ Sender = (*ResendSender)(nil)
)

type NullSender struct{}

func NewNullSender() *NullSender {
	return &NullSender{}
}

func (s *NullSender) SendWelcomeEmail(ctx context.Context, email string, name string) error {
	return nil
}

func (s *NullSender) SendPasswordResetEmail(ctx context.Context, email, baseURL, token string) error {
	return nil
}
