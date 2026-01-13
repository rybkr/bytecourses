package middleware

import (
	"bytecourses/internal/domain"
	"context"
)

type ctxKey int

const (
	ctxUserKey ctxKey = iota
	ctxProposalKey
	ctxCourseKey
)

func withUser(ctx context.Context, u *domain.User) context.Context {
	return context.WithValue(ctx, ctxUserKey, u)
}

func UserFromContext(ctx context.Context) (*domain.User, bool) {
	u, ok := ctx.Value(ctxUserKey).(*domain.User)
	return u, ok
}

func withProposal(ctx context.Context, p *domain.Proposal) context.Context {
	return context.WithValue(ctx, ctxProposalKey, p)
}

func ProposalFromContext(ctx context.Context) (*domain.Proposal, bool) {
	p, ok := ctx.Value(ctxProposalKey).(*domain.Proposal)
	return p, ok
}

func withCourse(ctx context.Context, c *domain.Course) context.Context {
	return context.WithValue(ctx, ctxCourseKey, c)
}

func CourseFromContext(ctx context.Context) (*domain.Course, bool) {
	c, ok := ctx.Value(ctxCourseKey).(*domain.Course)
	return c, ok
}
