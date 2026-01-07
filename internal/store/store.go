package store

import (
	"bytecourses/internal/domain"
	"context"
)

type UserStore interface {
	CreateUser(context.Context, *domain.User) error
	GetUserByID(context.Context, int64) (*domain.User, bool)
	GetUserByEmail(context.Context, string) (*domain.User, bool)
	UpdateUser(context.Context, *domain.User) error
}

type ProposalStore interface {
	CreateProposal(context.Context, *domain.Proposal) error
	GetProposalByID(context.Context, int64) (*domain.Proposal, bool)
	ListProposalsByAuthorID(context.Context, int64) ([]domain.Proposal, error)
	ListAllSubmittedProposals(context.Context) ([]domain.Proposal, error)
	UpdateProposal(context.Context, *domain.Proposal) error
	DeleteProposalByID(context.Context, int64) error
}
