package services

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
)

type ProposalService struct {
	proposals store.ProposalStore
	users     store.UserStore
}

func NewProposalService(proposals store.ProposalStore, users store.UserStore) *ProposalService {
	return &ProposalService{
		proposals: proposals,
		users:     users,
	}
}

type CreateProposalRequest struct {
	AuthorID             int64  `json:"author_id"`
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	Outline              string `json:"outline"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (s *ProposalService) CreateProposal(ctx context.Context, request CreateProposalRequest) (*domain.Proposal, error) {
	proposal := &domain.Proposal{
		AuthorID:             request.AuthorID,
		Title:                request.Title,
		Summary:              request.Summary,
		TargetAudience:       request.TargetAudience,
		LearningObjectives:   request.LearningObjectives,
		Outline:              request.Outline,
		AssumedPrerequisites: request.AssumedPrerequisites,
		Status:               domain.ProposalStatusDraft,
	}

	if err := s.proposals.CreateProposal(ctx, proposal); err != nil {
		return nil, err
	}

	return proposal, nil
}

func (s *ProposalService) GetProposal(ctx context.Context, proposalID int64, userID int64) (*domain.Proposal, error) {
	proposal, ok := s.proposals.GetProposalByID(ctx, proposalID)
	if !ok {
		return nil, ErrNotFound
	}

	user, ok := s.users.GetUserByID(ctx, userID)
	if !ok {
		return nil, ErrUnauthorized
	}
	if !proposal.IsViewableBy(user) {
		return nil, ErrNotFound
	}

	return proposal, nil
}

type ListProposalsRequest struct {
	UserID int64
	Role   domain.UserRole
}

// ListProposals returns proposals based on user role
func (s *ProposalService) ListProposals(ctx context.Context, request ListProposalsRequest) ([]domain.Proposal, error) {
	if request.Role == domain.UserRoleAdmin {
		return s.proposals.ListAllSubmittedProposals(ctx)
	}
	return s.proposals.ListProposalsByAuthorID(ctx, request.UserID)
}

// UpdateProposalRequest represents proposal update input
type UpdateProposalRequest struct {
	ProposalID           int64
	UserID               int64
	Title                string
	Summary              string
	TargetAudience       string
	LearningObjectives   string
	Outline              string
	AssumedPrerequisites string
}

// UpdateProposal updates a proposal (only if user owns it and it's amendable)
func (s *ProposalService) UpdateProposal(ctx context.Context, req UpdateProposalRequest) error {
	proposal, ok := s.proposals.GetProposalByID(ctx, req.ProposalID)
	if !ok {
		return ErrNotFound
	}

	user, ok := s.users.GetUserByID(ctx, req.UserID)
	if !ok {
		return ErrUnauthorized
	}

	if !proposal.IsOwnedBy(user) {
		return ErrNotFound
	}

	if !proposal.IsAmendable() {
		return ErrConflict
	}

	proposal.Title = req.Title
	proposal.Summary = req.Summary
	proposal.TargetAudience = req.TargetAudience
	proposal.LearningObjectives = req.LearningObjectives
	proposal.Outline = req.Outline
	proposal.AssumedPrerequisites = req.AssumedPrerequisites

	return s.proposals.UpdateProposal(ctx, proposal)
}

// DeleteProposal deletes a proposal (only if user owns it)
func (s *ProposalService) DeleteProposal(ctx context.Context, proposalID int64, userID int64) error {
	proposal, ok := s.proposals.GetProposalByID(ctx, proposalID)
	if !ok {
		return ErrNotFound
	}

	user, ok := s.users.GetUserByID(ctx, userID)
	if !ok {
		return ErrUnauthorized
	}

	if !proposal.IsOwnedBy(user) {
		return ErrNotFound
	}

	return s.proposals.DeleteProposalByID(ctx, proposalID)
}

// SubmitProposal transitions a proposal to submitted status
func (s *ProposalService) SubmitProposal(ctx context.Context, proposalID int64, userID int64) error {
	proposal, ok := s.proposals.GetProposalByID(ctx, proposalID)
	if !ok {
		return ErrNotFound
	}

	user, ok := s.users.GetUserByID(ctx, userID)
	if !ok {
		return ErrUnauthorized
	}

	if !proposal.IsOwnedBy(user) {
		return ErrNotFound
	}

	if !proposal.IsAmendable() {
		return ErrConflict
	}

	proposal.Status = domain.ProposalStatusSubmitted
	return s.proposals.UpdateProposal(ctx, proposal)
}

// WithdrawProposal transitions a proposal to withdrawn status
func (s *ProposalService) WithdrawProposal(ctx context.Context, proposalID int64, userID int64) error {
	proposal, ok := s.proposals.GetProposalByID(ctx, proposalID)
	if !ok {
		return ErrNotFound
	}

	user, ok := s.users.GetUserByID(ctx, userID)
	if !ok {
		return ErrUnauthorized
	}

	if !proposal.IsOwnedBy(user) {
		return ErrNotFound
	}

	if proposal.Status != domain.ProposalStatusSubmitted {
		return ErrConflict
	}

	proposal.Status = domain.ProposalStatusWithdrawn
	return s.proposals.UpdateProposal(ctx, proposal)
}

// ReviewProposalRequest represents admin review action input
type ReviewProposalRequest struct {
	ProposalID int64
	ReviewerID int64
	Action     string // "approve", "reject", "request-changes"
	Notes      string
}

// ReviewProposal performs admin review actions (approve, reject, request-changes)
func (s *ProposalService) ReviewProposal(ctx context.Context, req ReviewProposalRequest) error {
	proposal, ok := s.proposals.GetProposalByID(ctx, req.ProposalID)
	if !ok {
		return ErrNotFound
	}

	reviewer, ok := s.users.GetUserByID(ctx, req.ReviewerID)
	if !ok {
		return ErrUnauthorized
	}

	if !reviewer.IsAdmin() {
		return ErrForbidden
	}

	if proposal.Status != domain.ProposalStatusSubmitted {
		return ErrConflict
	}

	proposal.ReviewerID = &req.ReviewerID
	proposal.ReviewNotes = req.Notes

	switch req.Action {
	case "approve":
		proposal.Status = domain.ProposalStatusApproved
	case "reject":
		proposal.Status = domain.ProposalStatusRejected
	case "request-changes":
		proposal.Status = domain.ProposalStatusChangesRequested
	default:
		return ErrInvalidInput
	}

	return s.proposals.UpdateProposal(ctx, proposal)
}
