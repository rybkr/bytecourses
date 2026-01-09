package services

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
)

type ProposalService struct {
	proposals store.ProposalStore
}

func NewProposalService(proposals store.ProposalStore) *ProposalService {
	return &ProposalService{
		proposals: proposals,
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

func (r *CreateProposalRequest) Normalize() {
	if r.Title == "" {
		r.Title = "Untitled Proposal"
	}
}

func (r *CreateProposalRequest) IsValid() bool {
	return r.Title != "" && r.AuthorID > 0
}

type CreateProposalResult struct {
	ProposalID int64 `json:"id"`
}

func (s *ProposalService) CreateProposal(ctx context.Context, request *CreateProposalRequest) (*domain.Proposal, error) {
	request.Normalize()
	if !request.IsValid() {
		return nil, ErrInvalidInput
	}

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

func (s *ProposalService) GetProposal(ctx context.Context, u *domain.User, p *domain.Proposal) (*domain.Proposal, error) {
	if !p.IsViewableBy(u) {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *ProposalService) ListProposals(ctx context.Context, u *domain.User) ([]domain.Proposal, error) {
	if u.Role == domain.UserRoleAdmin {
		return s.proposals.ListAllSubmittedProposals(ctx)
	}
	return s.proposals.ListProposalsByAuthorID(ctx, u.ID)
}

func (s *ProposalService) ListMyProposals(ctx context.Context, u *domain.User) ([]domain.Proposal, error) {
	return s.proposals.ListProposalsByAuthorID(ctx, u.ID)
}

type UpdateProposalRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	Outline              string `json:"outline"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (s *ProposalService) UpdateProposal(ctx context.Context, proposal *domain.Proposal, user *domain.User, request *UpdateProposalRequest) error {
	if !proposal.IsOwnedBy(user) {
		return ErrNotFound
	}
	if !proposal.IsAmendable() {
		return ErrConflict
	}

	proposal.Title = request.Title
	proposal.Summary = request.Summary
	proposal.TargetAudience = request.TargetAudience
	proposal.LearningObjectives = request.LearningObjectives
	proposal.Outline = request.Outline
	proposal.AssumedPrerequisites = request.AssumedPrerequisites

	return s.proposals.UpdateProposal(ctx, proposal)
}

// DeleteProposal deletes a proposal (only if user owns it)
func (s *ProposalService) DeleteProposal(ctx context.Context, proposal *domain.Proposal, user *domain.User) error {
	if !proposal.IsOwnedBy(user) {
		return ErrNotFound
	}

	return s.proposals.DeleteProposalByID(ctx, proposal.ID)
}

// SubmitProposal transitions a proposal to submitted status
func (s *ProposalService) SubmitProposal(ctx context.Context, proposal *domain.Proposal, user *domain.User) error {
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
func (s *ProposalService) WithdrawProposal(ctx context.Context, proposal *domain.Proposal, user *domain.User) error {
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
	Action string // "approve", "reject", "request-changes"
	Notes  string
}

// ReviewProposal performs admin review actions (approve, reject, request-changes)
func (s *ProposalService) ReviewProposal(ctx context.Context, proposal *domain.Proposal, reviewer *domain.User, req ReviewProposalRequest) error {
	if !reviewer.IsAdmin() {
		return ErrForbidden
	}

	if proposal.Status != domain.ProposalStatusSubmitted {
		return ErrConflict
	}

	proposal.ReviewerID = &reviewer.ID
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
