package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
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
		return errors.New("resend: send failed")
	}
	return nil
}
