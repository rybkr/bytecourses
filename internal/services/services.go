package services

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/notify"
	"bytecourses/internal/store"
	"log/slog"
)

type Services struct {
	Auth      *AuthService
	Proposals *ProposalService
}

type Dependencies struct {
	UserStore          store.UserStore
	ProposalStore      store.ProposalStore
	PasswordResetStore store.PasswordResetStore
	SessionStore       auth.SessionStore
	EmailSender        notify.EmailSender
	Logger             *slog.Logger
}

func New(d Dependencies) *Services {
	return &Services{
		Auth:      NewAuthService(d.UserStore, d.SessionStore, d.PasswordResetStore, d.EmailSender, d.Logger),
		Proposals: NewProposalService(d.ProposalStore, d.Logger),
	}
}
