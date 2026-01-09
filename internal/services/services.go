package services

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/notify"
	"bytecourses/internal/store"
)

// Services aggregates all business services
type Services struct {
	Auth      *AuthService
	Users     *UserService
	Proposals *ProposalService
}

// Dependencies contains all infrastructure dependencies needed by services
type Dependencies struct {
	UserStore          store.UserStore
	ProposalStore      store.ProposalStore
	PasswordResetStore store.PasswordResetStore
	SessionStore       auth.SessionStore
	EmailSender        notify.EmailSender
}

// New creates and wires all services with their dependencies
func New(deps Dependencies) *Services {
	return &Services{
		Auth:      NewAuthService(deps.UserStore, deps.SessionStore, deps.PasswordResetStore, deps.EmailSender),
		Users:     NewUserService(deps.UserStore),
		Proposals: NewProposalService(deps.ProposalStore, deps.UserStore),
	}
}
