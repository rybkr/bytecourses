package services

import (
	"context"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/pkg/validation"
)

type ProposalService struct {
	Proposals persistence.ProposalRepository
	Users     persistence.UserRepository
	Events    events.EventBus
}

func NewProposalService(
	proposalRepo persistence.ProposalRepository,
	userRepo persistence.UserRepository,
	eventBus events.EventBus,
) *ProposalService {
	return &ProposalService{
		Proposals: proposalRepo,
		Users:     userRepo,
		Events:    eventBus,
	}
}

var (
	_ Command = (*CreateProposalCommand)(nil)
	_ Command = (*UpdateProposalCommand)(nil)
	_ Command = (*UpdateProposalStatusCommand)(nil)
	_ Command = (*DeleteProposalCommand)(nil)
)

type CreateProposalCommand struct {
	AuthorID             int64  `json:"author_id"`
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	Qualifications       string `json:"qualifications"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	Outline              string `json:"outline"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (c *CreateProposalCommand) Validate(v *validation.Validator) {
	v.Field(c.AuthorID, "author_id").EntityID()
	v.Field(c.Title, "title").MaxLength(128).IsTrimmed()
	v.Field(c.Summary, "summary").MaxLength(2048).IsTrimmed()
	v.Field(c.Qualifications, "qualifications").MaxLength(2048).IsTrimmed()
	v.Field(c.TargetAudience, "target_audience").MaxLength(2048).IsTrimmed()
	v.Field(c.LearningObjectives, "learning_objectives").MaxLength(2048).IsTrimmed()
	v.Field(c.Outline, "outline").MaxLength(2048).IsTrimmed()
	v.Field(c.AssumedPrerequisites, "assumed_prerequisites").MaxLength(2048).IsTrimmed()
}

func (s *ProposalService) Create(ctx context.Context, cmd *CreateProposalCommand) (*domain.Proposal, error) {
	if err := validation.Validate(cmd); err != nil {
		return nil, err
	}

	proposal := domain.Proposal{
		AuthorID:             cmd.AuthorID,
		Title:                cmd.Title,
		Summary:              cmd.Summary,
		Qualifications:       cmd.Qualifications,
		TargetAudience:       cmd.TargetAudience,
		LearningObjectives:   cmd.LearningObjectives,
		Outline:              cmd.Outline,
		AssumedPrerequisites: cmd.AssumedPrerequisites,
		Status:               domain.ProposalStatusDraft,
	}
	if err := s.Proposals.Create(ctx, &proposal); err != nil {
		return nil, err
	}

	event := domain.NewProposalCreatedEvent(proposal.ID, proposal.AuthorID)
	_ = s.Events.Publish(ctx, event)

	return &proposal, nil
}

type UpdateProposalCommand struct {
	ProposalID           int64  `json:"proposal_id"`
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	Qualifications       string `json:"qualifications"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	Outline              string `json:"outline"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
	UserID               int64  `json:"user_id"`
}

func (c *UpdateProposalCommand) Validate(v *validation.Validator) {
	v.Field(c.ProposalID, "proposal_id").EntityID()
	v.Field(c.Title, "title").MaxLength(128).IsTrimmed()
	v.Field(c.Summary, "summary").MaxLength(2048).IsTrimmed()
	v.Field(c.Qualifications, "qualifications").MaxLength(2048).IsTrimmed()
	v.Field(c.TargetAudience, "target_audience").MaxLength(2048).IsTrimmed()
	v.Field(c.LearningObjectives, "learning_objectives").MaxLength(2048).IsTrimmed()
	v.Field(c.Outline, "outline").MaxLength(2048).IsTrimmed()
	v.Field(c.AssumedPrerequisites, "assumed_prerequisites").MaxLength(2048).IsTrimmed()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ProposalService) Update(ctx context.Context, cmd *UpdateProposalCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	proposal, ok := s.Proposals.GetByID(ctx, cmd.ProposalID)
	if !ok {
		return errors.ErrNotFound
	}
	if !proposal.IsAmendable() {
		return errors.ErrInvalidStatusTransition
	}
	if proposal.AuthorID != cmd.UserID {
		return errors.ErrForbidden
	}

	proposal.Title = cmd.Title
	proposal.Summary = cmd.Summary
	proposal.Qualifications = cmd.Qualifications
	proposal.TargetAudience = cmd.TargetAudience
	proposal.LearningObjectives = cmd.LearningObjectives
	proposal.Outline = cmd.Outline
	proposal.AssumedPrerequisites = cmd.AssumedPrerequisites
	if err := s.Proposals.Update(ctx, proposal); err != nil {
		return err
	}

	event := domain.NewProposalUpdatedEvent(proposal.ID, proposal.AuthorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type UpdateProposalStatusCommand struct {
	ProposalID  int64                 `json:"proposal_id`
	Status      domain.ProposalStatus `json:"status"`
	ReviewNotes string                `json:"review_notes"`
	UserID      int64                 `json:"user_id"`
	UserRole    domain.UserRole       `json:"user_role"`
}

func (c *UpdateProposalStatusCommand) Validate(v *validation.Validator) {
	v.Field(c.ProposalID, "proposal_id").EntityID()
	v.Field(c.Status, "status").Required().IsTrimmed()
	if c.Status != domain.ProposalStatusSubmitted &&
		c.Status != domain.ProposalStatusWithdrawn &&
		c.Status != domain.ProposalStatusApproved &&
		c.Status != domain.ProposalStatusRejected &&
		c.Status != domain.ProposalStatusChangesRequested {
		v.Field("", "status").Required()
	}
	v.Field(c.ReviewNotes, "review_notes").IsTrimmed()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ProposalService) UpdateStatus(ctx context.Context, cmd *UpdateProposalStatusCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	proposal, ok := s.Proposals.GetByID(ctx, cmd.ProposalID)
	if !ok {
		return errors.ErrNotFound
	}

	switch cmd.Status {
	case domain.ProposalStatusSubmitted,
		domain.ProposalStatusWithdrawn:
		if proposal.AuthorID != cmd.UserID {
			return errors.ErrForbidden
		}
		proposal.Status = cmd.Status

	case domain.ProposalStatusApproved,
		domain.ProposalStatusRejected,
		domain.ProposalStatusChangesRequested:
		if cmd.UserRole != domain.UserRoleAdmin {
			return errors.ErrForbidden
		}
		proposal.Status = cmd.Status
		proposal.ReviewerID = &cmd.UserID
		proposal.ReviewNotes = cmd.ReviewNotes
	}

	if err := s.Proposals.Update(ctx, proposal); err != nil {
		return err
	}

	var event domain.Event
	switch cmd.Status {
	case domain.ProposalStatusSubmitted:
		event = domain.NewProposalSubmittedEvent(cmd.ProposalID, cmd.UserID, proposal.Title)
	case domain.ProposalStatusWithdrawn:
		event = domain.NewProposalWithdrawnEvent(cmd.ProposalID, cmd.UserID)
	case domain.ProposalStatusApproved:
		event = domain.NewProposalApprovedEvent(cmd.ProposalID, proposal.AuthorID, cmd.UserID, proposal.Title)
	case domain.ProposalStatusRejected:
		event = domain.NewProposalRejectedEvent(cmd.ProposalID, proposal.AuthorID, cmd.UserID, proposal.Title)
	case domain.ProposalStatusChangesRequested:
		event = domain.NewProposalChangesRequestedEvent(cmd.ProposalID, proposal.AuthorID, cmd.UserID, proposal.Title)
	}

	_ = s.Events.Publish(ctx, event)

	return nil
}

type DeleteProposalCommand struct {
	ProposalID int64 `json:"proposal_id"`
	UserID     int64 `json:"user_id"`
}

func (c *DeleteProposalCommand) Validate(v *validation.Validator) {
	v.Field(c.ProposalID, "proposal_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *ProposalService) Delete(ctx context.Context, cmd *DeleteProposalCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	proposal, ok := s.Proposals.GetByID(ctx, cmd.ProposalID)
	if !ok {
		return errors.ErrNotFound
	}
	if proposal.AuthorID != cmd.UserID {
		return errors.ErrForbidden
	}

	if err := s.Proposals.DeleteByID(ctx, cmd.ProposalID); err != nil {
		return err
	}

	event := domain.NewProposalDeletedEvent(proposal.ID, proposal.AuthorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type GetProposalQuery struct {
	ProposalID int64           `json:"proposal_id"`
	UserID     int64           `json:"user_id"`
	UserRole   domain.UserRole `json:"user_role"`
}

func (s *ProposalService) Get(ctx context.Context, query *GetProposalQuery) (*domain.Proposal, error) {
	proposal, ok := s.Proposals.GetByID(ctx, query.ProposalID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	switch query.UserRole {
	case domain.UserRoleStudent,
		domain.UserRoleInstructor:
		if proposal.AuthorID != query.UserID {
			return nil, errors.ErrNotFound
		}

	case domain.UserRoleAdmin:
		if proposal.Status != domain.ProposalStatusSubmitted &&
			proposal.Status != domain.ProposalStatusApproved &&
			proposal.Status != domain.ProposalStatusRejected &&
			proposal.Status != domain.ProposalStatusChangesRequested {
			return nil, errors.ErrNotFound
		}

	default:
		return nil, errors.ErrForbidden
	}

	return proposal, nil
}

type ListProposalsQuery struct {
	UserID   int64           `json:"user_id"`
	UserRole domain.UserRole `json:"user_role"`
}

func (s *ProposalService) List(ctx context.Context, query *ListProposalsQuery) ([]domain.Proposal, error) {
	proposals := make([]domain.Proposal, 0)
	var err error

	switch query.UserRole {
	case domain.UserRoleStudent,
		domain.UserRoleInstructor:
		proposals, err = s.Proposals.ListByAuthorID(ctx, query.UserID)

	case domain.UserRoleAdmin:
		proposals, err = s.Proposals.ListAllSubmitted(ctx)

	default:
		return make([]domain.Proposal, 0), errors.ErrForbidden
	}

	return proposals, err
}
