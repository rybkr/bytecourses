package services

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"log/slog"
)

type ProposalService struct {
	proposals store.ProposalStore
	logger    *ProposalLogger
}

func NewProposalService(proposals store.ProposalStore, logger *slog.Logger) *ProposalService {
	return &ProposalService{
		proposals: proposals,
		logger:    NewProposalLogger(logger),
	}
}

type CreateProposalRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	Qualifications       string `json:"qualifications"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	Outline              string `json:"outline"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
	AuthorID             int64  `json:"author_id"`
}

func (r *CreateProposalRequest) IsValid() bool {
	return r.AuthorID > 0
}

func (s *ProposalService) CreateProposal(ctx context.Context, request *CreateProposalRequest) (*domain.Proposal, error) {
	if !request.IsValid() {
		return nil, ErrInvalidInput
	}

	proposal := &domain.Proposal{
		AuthorID:             request.AuthorID,
		Title:                request.Title,
		Summary:              request.Summary,
		Qualifications:       request.Qualifications,
		TargetAudience:       request.TargetAudience,
		LearningObjectives:   request.LearningObjectives,
		Outline:              request.Outline,
		AssumedPrerequisites: request.AssumedPrerequisites,
		Status:               domain.ProposalStatusDraft,
	}
	if err := s.proposals.CreateProposal(ctx, proposal); err != nil {
		s.logger.Error("proposal creation failed",
			"event", "proposal.creation",
			"user_id", request.AuthorID,
			"title", request.Title,
			"error", err,
		)
		return nil, err
	}

	s.logger.Info("proposal.created",
		"proposal_id", proposal.ID,
		"user_id", request.AuthorID,
		"title", proposal.Title,
		"status", proposal.Status,
	)

	return proposal, nil
}

func (s *ProposalService) GetProposal(ctx context.Context, p *domain.Proposal, u *domain.User) (*domain.Proposal, error) {
	if !p.IsViewableBy(u) {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *ProposalService) ListProposals(ctx context.Context, u *domain.User) ([]domain.Proposal, error) {
	if u.IsAdmin() {
		return s.proposals.ListAllSubmittedProposals(ctx)
	} else {
		return s.proposals.ListProposalsByAuthorID(ctx, u.ID)
	}
}

func (s *ProposalService) ListMyProposals(ctx context.Context, u *domain.User) ([]domain.Proposal, error) {
	return s.proposals.ListProposalsByAuthorID(ctx, u.ID)
}

type UpdateProposalRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	Qualifications       string `json:"qualifications"`
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
	proposal.Qualifications = request.Qualifications
	proposal.TargetAudience = request.TargetAudience
	proposal.LearningObjectives = request.LearningObjectives
	proposal.Outline = request.Outline
	proposal.AssumedPrerequisites = request.AssumedPrerequisites

	err := s.proposals.UpdateProposal(ctx, proposal)
	if err != nil {
		s.logger.ErrorOp("update", proposal, user, err)
		return err
	}

	s.logger.Info("proposal.updated",
		"proposal_id", proposal.ID,
		"user_id", user.ID,
		"status", proposal.Status,
	)

	return nil
}

func (s *ProposalService) DeleteProposal(ctx context.Context, proposal *domain.Proposal, user *domain.User) error {
	if !proposal.IsOwnedBy(user) {
		return ErrNotFound
	}

	err := s.proposals.DeleteProposalByID(ctx, proposal.ID)
	if err != nil {
		s.logger.ErrorOp("delete", proposal, user, err)
		return err
	}

	s.logger.Info("proposal.deleted",
		"proposal_id", proposal.ID,
		"user_id", user.ID,
		"title", proposal.Title,
		"status", proposal.Status,
	)

	return nil
}

type ProposalActionRequest struct {
	ReviewNotes string `json:"review_notes"`
}

func (s *ProposalService) SubmitProposal(ctx context.Context, proposal *domain.Proposal, user *domain.User) error {
	if !proposal.IsOwnedBy(user) {
		return ErrNotFound
	}
	if !proposal.IsAmendable() {
		return ErrConflict
	}

	oldStatus := proposal.Status
	proposal.Status = domain.ProposalStatusSubmitted
	err := s.proposals.UpdateProposal(ctx, proposal)
	if err != nil {
		s.logger.Error("proposal submission failed",
			"event", "proposal.submission",
			"proposal_id", proposal.ID,
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	s.logger.InfoTransition("submit", proposal, user, oldStatus, proposal.Status)
	return nil
}

func (s *ProposalService) WithdrawProposal(ctx context.Context, proposal *domain.Proposal, user *domain.User) error {
	if !proposal.IsOwnedBy(user) {
		return ErrNotFound
	}
	if proposal.Status != domain.ProposalStatusSubmitted {
		return ErrConflict
	}

	oldStatus := proposal.Status
	proposal.Status = domain.ProposalStatusWithdrawn
	err := s.proposals.UpdateProposal(ctx, proposal)
	if err != nil {
		s.logger.Error("proposal withdrawal failed",
			"event", "proposal.withdrawal",
			"proposal_id", proposal.ID,
			"user_id", user.ID,
			"error", err,
		)
		return err
	}

	s.logger.InfoTransition("withdraw", proposal, user, oldStatus, proposal.Status)
	return nil
}

type ReviewProposalRequest struct {
	Action string
	Notes  string
}

func (r *ReviewProposalRequest) IsValid() bool {
	return r.Action == "approve" || r.Action == "reject" || r.Action == "request-changes"
}

func (s *ProposalService) ReviewProposal(ctx context.Context, proposal *domain.Proposal, reviewer *domain.User, request *ReviewProposalRequest) error {
	if !reviewer.IsAdmin() {
		return ErrForbidden
	}
	if proposal.Status != domain.ProposalStatusSubmitted {
		return ErrConflict
	}
	if !request.IsValid() {
		return ErrInvalidInput
	}

	oldStatus := proposal.Status
	proposal.ReviewerID = &reviewer.ID
	proposal.ReviewNotes = request.Notes

	switch request.Action {
	case "approve":
		proposal.Status = domain.ProposalStatusApproved
	case "reject":
		proposal.Status = domain.ProposalStatusRejected
	case "request-changes":
		proposal.Status = domain.ProposalStatusChangesRequested
	default:
		return ErrInvalidInput
	}

	err := s.proposals.UpdateProposal(ctx, proposal)
	if err != nil {
		s.logger.Error("proposal review failed",
			"event", "proposal.review",
			"proposal_id", proposal.ID,
			"reviewer_id", reviewer.ID,
			"action", request.Action,
			"error", err,
		)
		return err
	}

	s.logger.Info("proposal.reviewed",
		"proposal_id", proposal.ID,
		"reviewer_id", reviewer.ID,
		"action", request.Action,
		"old_status", oldStatus,
		"new_status", proposal.Status,
		"has_notes", request.Notes != "",
	)
	return nil
}
