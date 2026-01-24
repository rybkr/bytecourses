package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type ResendSender struct {
	apiKey    string
	fromEmail string
	client    *http.Client
}

func NewResendSender(apiKey, fromEmail string) *ResendSender {
	return &ResendSender{
		apiKey:    apiKey,
		fromEmail: fromEmail,
		client:    &http.Client{},
	}
}

var (
	_ Sender = (*ResendSender)(nil)
)

type resendRequest struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

func (s *ResendSender) sendEmail(ctx context.Context, to, subject, html string) error {
	reqBody := resendRequest{
		From:    s.fromEmail,
		To:      to,
		Subject: subject,
		HTML:    html,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API error: status %d", resp.StatusCode)
	}

	return nil
}

func (s *ResendSender) SendWelcomeEmail(ctx context.Context, email, name, getStartedURL string) error {
	subject := "Welcome to ByteCourses!"

	var buf bytes.Buffer
	data := struct {
		Name          string
		GetStartedURL string
	}{Name: name, GetStartedURL: getStartedURL}
	if err := welcomeTemplate.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute welcome template: %w", err)
	}

	return s.sendEmail(ctx, email, subject, buf.String())
}

func (s *ResendSender) SendPasswordResetEmail(ctx context.Context, email, resetURL, token string) error {
	subject := "Reset Your Password"

	u, err := url.Parse(resetURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("resend: invalid base url %s", resetURL)
	}

	query := u.Query()
	query.Set("token", token)
	query.Set("email", email)
	u.RawQuery = query.Encode()
	resetURL = u.String()

	var buf bytes.Buffer
	data := struct{ ResetURL string }{ResetURL: resetURL}
	if err := passwordResetTemplate.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute password reset template: %w", err)
	}

	return s.sendEmail(ctx, email, subject, buf.String())
}

func (s *ResendSender) SendProposalSubmittedEmail(ctx context.Context, email, name, title, proposalURL string) error {
	subject := "Proposal Submitted"
	var buf bytes.Buffer
	data := struct {
		Name          string
		ProposalTitle string
		ProposalURL   string
	}{Name: name, ProposalTitle: title, ProposalURL: proposalURL}
	if err := proposalSubmittedTemplate.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute proposal submitted template: %w", err)
	}
	return s.sendEmail(ctx, email, subject, buf.String())
}

func (s *ResendSender) SendProposalApprovedEmail(ctx context.Context, email, name, title, courseURL string) error {
	subject := "Proposal Approved"
	var buf bytes.Buffer
	data := struct {
		Name          string
		ProposalTitle string
		CourseURL     string
	}{Name: name, ProposalTitle: title, CourseURL: courseURL}
	if err := proposalApprovedTemplate.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute proposal approved template: %w", err)
	}
	return s.sendEmail(ctx, email, subject, buf.String())
}

func (s *ResendSender) SendProposalRejectedEmail(ctx context.Context, email, name, title, reviewNotes, newProposalURL string) error {
	subject := "Proposal Not Approved"
	var buf bytes.Buffer
	data := struct {
		Name           string
		ProposalTitle  string
		ReviewNotes    string
		NewProposalURL string
	}{Name: name, ProposalTitle: title, ReviewNotes: reviewNotes, NewProposalURL: newProposalURL}
	if err := proposalRejectedTemplate.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute proposal rejected template: %w", err)
	}
	return s.sendEmail(ctx, email, subject, buf.String())
}

func (s *ResendSender) SendProposalChangesRequestedEmail(ctx context.Context, email, name, title, reviewNotes, proposalURL string) error {
	subject := "Changes Requested"
	var buf bytes.Buffer
	data := struct {
		Name          string
		ProposalTitle string
		ReviewNotes   string
		ProposalURL   string
	}{Name: name, ProposalTitle: title, ReviewNotes: reviewNotes, ProposalURL: proposalURL}
	if err := proposalChangesTemplate.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute proposal changes requested template: %w", err)
	}
	return s.sendEmail(ctx, email, subject, buf.String())
}
