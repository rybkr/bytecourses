package notify

import (
	"context"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, text, html string) error
}
