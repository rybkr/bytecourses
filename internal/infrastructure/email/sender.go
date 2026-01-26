package email

import (
	"context"
)

type Sender interface {
	SendWelcomeEmail(ctx context.Context, email, name, getStartedURL string) error
	SendPasswordResetEmail(ctx context.Context, email, baseURL, token string) error
	SendProposalSubmittedEmail(ctx context.Context, email, name, title, proposalURL string) error
	SendProposalApprovedEmail(ctx context.Context, email, name, title, courseURL string) error
	SendProposalRejectedEmail(ctx context.Context, email, name, title, reviewNotes, newProposalURL string) error
	SendProposalChangesRequestedEmail(ctx context.Context, email, name, title, reviewNotes, proposalURL string) error
	SendEnrollmentConfirmationEmail(ctx context.Context, email, name, courseTitle, courseURL string) error
}

var (
	_ Sender = (*NullSender)(nil)
	_ Sender = (*ResendSender)(nil)
)

type NullSender struct{}

func NewNullSender() *NullSender {
	return &NullSender{}
}

func (s *NullSender) SendWelcomeEmail(ctx context.Context, email, name, getStartedURL string) error {
	return nil
}

func (s *NullSender) SendPasswordResetEmail(ctx context.Context, email, baseURL, token string) error {
	return nil
}

func (s *NullSender) SendProposalSubmittedEmail(ctx context.Context, email, name, title, proposalURL string) error {
	return nil
}

func (s *NullSender) SendProposalApprovedEmail(ctx context.Context, email, name, title, courseURL string) error {
	return nil
}

func (s *NullSender) SendProposalRejectedEmail(ctx context.Context, email, name, title, reviewNotes, newProposalURL string) error {
	return nil
}

func (s *NullSender) SendProposalChangesRequestedEmail(ctx context.Context, email, name, title, reviewNotes, proposalURL string) error {
	return nil
}

func (s *NullSender) SendEnrollmentConfirmationEmail(ctx context.Context, email, name, courseTitle, courseURL string) error {
	return nil
}
