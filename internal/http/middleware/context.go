package middleware

import (
	"bytecourses/internal/domain"
	"context"
)

type ctxKey int

const (
	ctxUserKey ctxKey = iota
	ctxProposalKey
)

func withUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, ctxUserKey, u)
}

// UserFromContext retrieves the authenticated user injected by auth middleware.
// It returns false only if middleware was not applied or wiring is broken.
func UserFromContext(ctx context.Context) (*domain.User, bool) {
	u, ok := ctx.Value(ctxUserKey).(*domain.User)
	return u, ok
}

func withProposal(ctx context.Context, p *domain.Proposal) context.Context {
	return context.WithValue(ctx, ctxProposalKey, p)
}

// ProposalFromContext retrieves the proposal loaded by proposal middleware.
// It returns false only if middleware was not applied or wiring is broken.
func ProposalFromContext(ctx context.Context) (*domain.Proposal, bool) {
	p, ok := ctx.Value(ctxProposalKey).(*domain.Proposal)
	return p, ok
}
