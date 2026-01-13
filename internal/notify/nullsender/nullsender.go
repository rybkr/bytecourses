package nullsender

import (
	"context"
)

type Sender struct{}

func New() *Sender {
	return &Sender{}
}

func (s *Sender) Send(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (s *Sender) SendPasswordResetPrompt(_ context.Context, _, _, _ string) error {
	return nil
}

func (s *Sender) SendWelcomeEmail(_ context.Context, _, _ string) error {
	return nil
}
