package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

func (s *ResendSender) SendWelcomeEmail(ctx context.Context, email, name string) error {
	subject := "Welcome to Bytecourses!"
	html := fmt.Sprintf(`
		<h1>Welcome, %s!</h1>
		<p>Thank you for joining Bytecourses. We're excited to have you!</p>
		<p>You can now start creating and submitting course proposals.</p>
	`, name)

	return s.sendEmail(ctx, email, subject, html)
}

func (s *ResendSender) SendPasswordResetEmail(ctx context.Context, email string) error {
	subject := "Reset Your Password"
	html := `
		<h1>Password Reset Request</h1>
		<p>We received a request to reset your password.</p>
		<p>If you didn't make this request, you can ignore this email.</p>
		<p>This link will expire in 1 hour.</p>
	`

	return s.sendEmail(ctx, email, subject, html)
}
