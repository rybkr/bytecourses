package services

import (
	"context"
	"strings"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/pkg/validation"
)

// ProposalService handles all proposal operations.
type ProposalService struct {
	proposals persistence.ProposalRepository
	users     persistence.UserRepository
	events    events.EventBus
}

// NewProposalService creates a new ProposalService with the given dependencies.
func NewProposalService(
	proposals persistence.ProposalRepository,
	users persistence.UserRepository,
	eventBus events.EventBus,
) *ProposalService {
	return &ProposalService{
		proposals: proposals,
		users:     users,
		events:    eventBus,
	}
}

// CreateProposalInput contains the data needed to create a proposal.
type CreateProposalInput struct {
	AuthorID             int64
	Title                string
	Summary              string
	Qualifications       string
	TargetAudience       string
	LearningObjectives   string
	Outline              string
	AssumedPrerequisites string
}

func (i *CreateProposalInput) Validate(v *validation.Validator) {
	v.Field(i.AuthorID, "author_id").EntityID()
	v.Field(i.Title, "title").Required().MinLength(4).MaxLength(128)
	v.Field(i.Summary, "summary").Required().MaxLength(2048)
	v.Field(i.Qualifications, "qualifications").MaxLength(2048)
	v.Field(i.TargetAudience, "target_audience").MaxLength(2048)
	v.Field(i.LearningObjectives, "learning_objectives").MaxLength(2048)
	v.Field(i.Outline, "outline").MaxLength(2048)
	v.Field(i.AssumedPrerequisites, "assumed_prerequisites").MaxLength(2048)
}

// Create creates a new proposal.
func (s *ProposalService) Create(ctx context.Context, input *CreateProposalInput) (*domain.Proposal, error) {
	if err := validation.New().Validate(input); err != nil {
		return nil, err
	}

	proposal := &domain.Proposal{
		AuthorID:             input.AuthorID,
		Title:                strings.TrimSpace(input.Title),
		Summary:              strings.TrimSpace(input.Summary),
		Qualifications:       strings.TrimSpace(input.Qualifications),
		TargetAudience:       strings.TrimSpace(input.TargetAudience),
		LearningObjectives:   strings.TrimSpace(input.LearningObjectives),
		Outline:              strings.TrimSpace(input.Outline),
		AssumedPrerequisites: strings.TrimSpace(input.AssumedPrerequisites),
		Status:               domain.ProposalStatusDraft,
	}

	if err := s.proposals.Create(ctx, proposal); err != nil {
		return nil, err
	}

	event := domain.NewProposalCreatedEvent(proposal.ID, proposal.AuthorID)
	_ = s.events.Publish(ctx, event)

	return proposal, nil
}

// UpdateProposalInput contains the data needed to update a proposal.
type UpdateProposalInput struct {
	ProposalID           int64
	UserID               int64
	Title                string
	Summary              string
	Qualifications       string
	TargetAudience       string
	LearningObjectives   string
	Outline              string
	AssumedPrerequisites string
}

func (i *UpdateProposalInput) Validate(v *validation.Validator) {
	v.Field(i.ProposalID, "proposal_id").EntityID()
	v.Field(i.UserID, "user_id").EntityID()
	v.Field(i.Title, "title").Required().MinLength(4).MaxLength(128)
	v.Field(i.Summary, "summary").Required().MaxLength(2048)
	v.Field(i.Qualifications, "qualifications").MaxLength(2048)
	v.Field(i.TargetAudience, "target_audience").MaxLength(2048)
	v.Field(i.LearningObjectives, "learning_objectives").MaxLength(2048)
	v.Field(i.Outline, "outline").MaxLength(2048)
	v.Field(i.AssumedPrerequisites, "assumed_prerequisites").MaxLength(2048)
}

// Update updates an existing proposal.
func (s *ProposalService) Update(ctx context.Context, input *UpdateProposalInput) (*domain.Proposal, error) {
	if err := validation.New().Validate(input); err != nil {
		return nil, err
	}

	proposal, ok := s.proposals.GetByID(ctx, input.ProposalID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if proposal.AuthorID != input.UserID {
		return nil, errors.ErrForbidden
	}

	if !proposal.IsAmendable() {
		return nil, errors.ErrInvalidStatusTransition
	}

	proposal.Title = strings.TrimSpace(input.Title)
	proposal.Summary = strings.TrimSpace(input.Summary)
	proposal.Qualifications = strings.TrimSpace(input.Qualifications)
	proposal.TargetAudience = strings.TrimSpace(input.TargetAudience)
	proposal.LearningObjectives = strings.TrimSpace(input.LearningObjectives)
	proposal.Outline = strings.TrimSpace(input.Outline)
	proposal.AssumedPrerequisites = strings.TrimSpace(input.AssumedPrerequisites)

	if err := s.proposals.Update(ctx, proposal); err != nil {
		return nil, err
	}

	event := domain.NewProposalUpdatedEvent(proposal.ID, proposal.AuthorID)
	_ = s.events.Publish(ctx, event)

	return proposal, nil
}

// SubmitProposalInput contains the data needed to submit a proposal.
type SubmitProposalInput struct {
	ProposalID int64
	UserID     int64
}

// Submit submits a proposal for review.
func (s *ProposalService) Submit(ctx context.Context, input *SubmitProposalInput) (*domain.Proposal, error) {
	proposal, ok := s.proposals.GetByID(ctx, input.ProposalID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if proposal.AuthorID != input.UserID {
		return nil, errors.ErrForbidden
	}

	if !proposal.IsAmendable() {
		return nil, errors.ErrInvalidStatusTransition
	}

	oldStatus := proposal.Status
	proposal.Status = domain.ProposalStatusSubmitted

	if err := s.proposals.Update(ctx, proposal); err != nil {
		return nil, err
	}

	event := domain.NewProposalSubmittedEvent(proposal.ID, proposal.AuthorID, proposal.Title, oldStatus)
	_ = s.events.Publish(ctx, event)

	return proposal, nil
}

// WithdrawProposalInput contains the data needed to withdraw a proposal.
type WithdrawProposalInput struct {
	ProposalID int64
	UserID     int64
}

// Withdraw withdraws a submitted proposal.
func (s *ProposalService) Withdraw(ctx context.Context, input *WithdrawProposalInput) (*domain.Proposal, error) {
	proposal, ok := s.proposals.GetByID(ctx, input.ProposalID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if proposal.AuthorID != input.UserID {
		return nil, errors.ErrForbidden
	}

	if proposal.Status != domain.ProposalStatusSubmitted {
		return nil, errors.ErrInvalidStatusTransition
	}

	proposal.Status = domain.ProposalStatusWithdrawn

	if err := s.proposals.Update(ctx, proposal); err != nil {
		return nil, err
	}

	event := domain.NewProposalWithdrawnEvent(proposal.ID, proposal.AuthorID)
	_ = s.events.Publish(ctx, event)

	return proposal, nil
}

// ReviewDecision represents a review decision.
type ReviewDecision string

const (
	ReviewDecisionApprove        ReviewDecision = "approve"
	ReviewDecisionReject         ReviewDecision = "reject"
	ReviewDecisionRequestChanges ReviewDecision = "request_changes"
)

// ReviewProposalInput contains the data needed to review a proposal.
type ReviewProposalInput struct {
	ProposalID int64
	ReviewerID int64
	Decision   ReviewDecision
	Notes      string
}

func (i *ReviewProposalInput) Validate(v *validation.Validator) {
	v.Field(i.ProposalID, "proposal_id").EntityID()
	v.Field(i.ReviewerID, "reviewer_id").EntityID()
	if i.Decision != ReviewDecisionApprove && i.Decision != ReviewDecisionReject && i.Decision != ReviewDecisionRequestChanges {
		v.Field("", "decision").Required()
	}
}

// Review reviews a submitted proposal.
func (s *ProposalService) Review(ctx context.Context, input *ReviewProposalInput) (*domain.Proposal, error) {
	if err := validation.New().Validate(input); err != nil {
		return nil, err
	}

	proposal, ok := s.proposals.GetByID(ctx, input.ProposalID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if proposal.Status != domain.ProposalStatusSubmitted {
		return nil, errors.ErrInvalidStatusTransition
	}

	proposal.ReviewerID = &input.ReviewerID
	proposal.ReviewNotes = input.Notes

	var event domain.Event
	switch input.Decision {
	case ReviewDecisionApprove:
		proposal.Status = domain.ProposalStatusApproved
		event = domain.NewProposalApprovedEvent(proposal.ID, proposal.AuthorID, input.ReviewerID, proposal.Title)
	case ReviewDecisionReject:
		proposal.Status = domain.ProposalStatusRejected
		event = domain.NewProposalRejectedEvent(proposal.ID, proposal.AuthorID, input.ReviewerID, proposal.Title)
	case ReviewDecisionRequestChanges:
		proposal.Status = domain.ProposalStatusChangesRequested
		event = domain.NewProposalChangesRequestedEvent(proposal.ID, proposal.AuthorID, input.ReviewerID, proposal.Title)
	}

	if err := s.proposals.Update(ctx, proposal); err != nil {
		return nil, err
	}

	_ = s.events.Publish(ctx, event)

	return proposal, nil
}

// DeleteProposalInput contains the data needed to delete a proposal.
type DeleteProposalInput struct {
	ProposalID int64
	UserID     int64
}

// Delete deletes a draft proposal.
func (s *ProposalService) Delete(ctx context.Context, input *DeleteProposalInput) error {
	proposal, ok := s.proposals.GetByID(ctx, input.ProposalID)
	if !ok {
		return errors.ErrNotFound
	}

	if proposal.AuthorID != input.UserID {
		return errors.ErrForbidden
	}

	if proposal.Status != domain.ProposalStatusDraft {
		return errors.ErrInvalidStatusTransition
	}

	if err := s.proposals.DeleteByID(ctx, input.ProposalID); err != nil {
		return err
	}

	event := domain.NewProposalDeletedEvent(proposal.ID, proposal.AuthorID, proposal.Status)
	_ = s.events.Publish(ctx, event)

	return nil
}

// GetByIDInput contains the data needed to get a proposal by ID.
type GetByIDInput struct {
	ProposalID int64
	UserID     int64
	IsAdmin    bool
}

// GetByID retrieves a proposal by ID with access control.
func (s *ProposalService) GetByID(ctx context.Context, input *GetByIDInput) (*domain.Proposal, error) {
	proposal, ok := s.proposals.GetByID(ctx, input.ProposalID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if proposal.AuthorID != input.UserID && !(input.IsAdmin && proposal.WasSubmitted()) {
		return nil, errors.ErrForbidden
	}

	return proposal, nil
}

// ProposalWithAuthor contains a proposal with its author information.
type ProposalWithAuthor struct {
	Proposal *domain.Proposal
	Author   *domain.User
}

// ListAll retrieves all submitted proposals (admin only).
func (s *ProposalService) ListAll(ctx context.Context, isAdmin bool) ([]ProposalWithAuthor, error) {
	if !isAdmin {
		return nil, errors.ErrForbidden
	}

	proposals, err := s.proposals.ListAllSubmitted(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ProposalWithAuthor, 0, len(proposals))
	for i := range proposals {
		p := &proposals[i]
		author, _ := s.users.GetByID(ctx, p.AuthorID)
		result = append(result, ProposalWithAuthor{
			Proposal: p,
			Author:   author,
		})
	}

	return result, nil
}

// ListMine retrieves all proposals for a specific user.
func (s *ProposalService) ListMine(ctx context.Context, userID int64) ([]domain.Proposal, error) {
	return s.proposals.ListByAuthorID(ctx, userID)
}
