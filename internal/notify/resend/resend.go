package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Sender struct {
	apiKey string
	from   string
	client *http.Client
}

func New(apiKey, from string) *Sender {
	return &Sender{
		apiKey: apiKey,
		from:   from,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *Sender) Send(ctx context.Context, to, subject, text, html string) error {
	if s.apiKey == "" || s.from == "" {
		return errors.New("resend: missing api key or from")
	}

	body := map[string]any{
		"from":    s.from,
		"to":      []string{to},
		"subject": subject,
	}
	if html != "" {
		body["html"] = html
	}
	if text != "" {
		body["text"] = text
	}

	b, _ := json.Marshal(body)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(b))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+s.apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		bodyBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("resend: send failed (status %d, body read error: %v)", response.StatusCode, err)
		}

		var errorResp struct {
			Message string `json:"message"`
			Name    string `json:"name"`
		}
		if err := json.Unmarshal(bodyBytes, &errorResp); err == nil && errorResp.Message != "" {
			return fmt.Errorf("resend: send failed (status %d): %s", response.StatusCode, errorResp.Message)
		}

		bodyStr := string(bodyBytes)
		if len(bodyStr) > 200 {
			bodyStr = bodyStr[:200] + "..."
		}
		return fmt.Errorf("resend: send failed (status %d): %s", response.StatusCode, bodyStr)
	}
	return nil
}

func (s *Sender) SendPasswordResetPrompt(ctx context.Context, to, resetURL, token string) error {
	if strings.TrimSpace(to) == "" {
		return errors.New("resend: missing recipient email")
	}
	if strings.TrimSpace(resetURL) == "" {
		return errors.New("resend: missing reset url")
	}

	u, err := url.Parse(resetURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("resend: invalid reset url")
	}

	q := u.Query()
	q.Set("token", token)
	q.Set("email", to)
	u.RawQuery = q.Encode()
	link := u.String()

	subject := "Reset your password"
	text := fmt.Sprintf(
		"Someone requested a password reset for %s.\n\nReset your password using this link:\n%s\n\nIf you didn't request this, you can ignore this email.\n",
		to, link,
	)
	html := fmt.Sprintf(
		`<p>Someone requested a password reset for <strong>%s</strong>.</p>
<p><a href="%s">Reset your password</a>.</p>
<p>If you didn't request this, you can ignore this email.</p>`,
		escapeHTML(to), escapeHTML(link),
	)

	return s.Send(ctx, to, subject, text, html)
}

func (s *Sender) SendWelcomeEmail(ctx context.Context, to, name string) error {
	if strings.TrimSpace(to) == "" {
		return errors.New("resend: missing recipient email")
	}

	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = "there"
	}

	subject := "Welcome to Byte Courses!"

	text := fmt.Sprintf(
		"Hi %s,\n\n"+
			"Welcome to Byte Courses! We're excited to have you join our community.\n\n"+
			"Byte Course is a platform where students can both learn and teach specialized topics through short courses. Whether you're here to:\n"+
			"- Learn from peer-taught courses on specialized topics\n"+
			"- Share your expertise by proposing and teaching your own course\n\n"+
			"We're here to support your educational journey.\n\n"+
			"If you have any questions, please direct them to help@bytecourses.org.\n\n"+
			"Happy learning!\n"+
			"The Byte Course Team",
		displayName,
	)

	html := fmt.Sprintf(
		`<p>Hi <strong>%s</strong>,</p>
<p>Welcome to <strong>Byte Courses</strong>! We're excited to have you join our community.</p>
<p>Byte Courses is a platform where students can both learn and teach specialized topics through short courses. Whether you're here to:</p>
<ul>
  <li>Learn from peer-taught courses on specialized topics</li>
  <li>Share your expertise by proposing and teaching your own course</li>
</ul>
<p>We're here to support your educational journey.</p>
<p>If you have any questions, please direct them to help@bytecourses.org.</p>
<p>Happy learning!<br>The Byte Course Team</p>`,
		escapeHTML(displayName),
	)

	return s.Send(ctx, to, subject, text, html)
}

func escapeHTML(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return r.Replace(s)
}
