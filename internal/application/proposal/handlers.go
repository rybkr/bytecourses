package proposal

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/pkg/validation"
	"bytecourses/internal/store"
	"context"
	"strings"
)

type CreateHandler struct {
	proposals store.ProposalStore
	eventBus  events.EventBus
	validator *validation.Validator
}

func NewCreateHandler(
	proposals store.ProposalStore,
	eventBus events.EventBus,
	validator *validation.Validator,
) *CreateHandler {
	return &CreateHandler{
		proposals: proposals,
		eventBus:  eventBus,
	}
}

func (h *CreateHandler) Handle(ctx context.Context, cmd *CreateCommand) (*domain.Proposal, error) {
	if err := h.validator.Validate(cmd); err != nil {
		return nil, err
	}

	proposal := &domain.Proposal{
		AuthorID:             cmd.AuthorID,
		Title:                strings.TrimSpace(cmd.Title),
		Summary:              strings.TrimSpace(cmd.Summary),
		Qualifications:       strings.TrimSpace(cmd.Qualifications),
		TargetAudience:       strings.TrimSpace(cmd.TargetAudience),
		LearningObjectives:   strings.TrimSpace(cmd.LearningObjectives),
		Outline:              strings.TrimSpace(cmd.Outline),
		AssumedPrerequisites: strings.TrimSpace(cmd.AssumedPrerequisites),
		Status:               domain.ProposalStatusDraft,
	}
	if err := h.proposals.CreateProposal(ctx, proposal); err != nil {
		return nil, err
	}

	event := domain.NewProposalCreatedEvent(proposal.ID, proposal.AuthorID, proposal.Title)
	_ = h.eventBus.Publish(ctx, event)

	return proposal, nil
}
