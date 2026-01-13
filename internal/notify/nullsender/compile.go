package nullsender

import (
	"bytecourses/internal/notify"
)

var _ notify.EmailSender = (*Sender)(nil)
