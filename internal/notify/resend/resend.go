package resend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
