package store

import (
	"bytecourses/internal/domain"
	"context"
)

type UserStore interface {
	InsertUser(context.Context, *domain.User) error
	GetUserByID(context.Context, int64) (domain.User, bool)
	GetUserByEmail(context.Context, string) (domain.User, bool)
	UpdateUser(context.Context, *domain.User) error
}

type ProposalStore interface {
	InsertProposal(context.Context, *domain.Proposal) error
	GetProposalByID(context.Context, int64) (domain.Proposal, bool)
	GetProposalsByUserID(context.Context, int64) []domain.Proposal
	UpdateProposal(context.Context, *domain.Proposal) error
	DeleteProposal(context.Context, int64) error
}
