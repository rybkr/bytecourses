package domain

import (
	"time"
)

type ProposalStatus string

const (
	ProposalStatusDraft            ProposalStatus = "draft"
	ProposalStatusSubmitted        ProposalStatus = "submitted"
	ProposalStatusWithdrawn        ProposalStatus = "withdrawn"
	ProposalStatusApproved         ProposalStatus = "approved"
	ProposalStatusRejected         ProposalStatus = "rejected"
	ProposalStatusChangesRequested ProposalStatus = "changes_requested"
)

type Proposal struct {
	ID                   int64          `json:"id"`
	Title                string         `json:"title"`
	Summary              string         `json:"summary"`
	AuthorID             int64          `json:"author_id"`
	TargetAudience       string         `json:"target_audience"`
	LearningObjectives   string         `json:"learning_objectives"`
	Outline              string         `json:"outline"`
	AssumedPrerequisites string         `json:"assumed_prerequisites"`
	ReviewNotes          string         `json:"review_notes"`
	ReviewerID           *int64         `json:"reviewer_id"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	Status               ProposalStatus `json:"status"`
}

func (p *Proposal) WasSubmitted() bool {
	return p.Status == ProposalStatusSubmitted ||
		p.Status == ProposalStatusApproved ||
		p.Status == ProposalStatusRejected ||
		p.Status == ProposalStatusChangesRequested
}

func (p *Proposal) IsOwnedBy(u *User) bool {
	return p.AuthorID == u.ID
}

func (p *Proposal) IsViewableBy(u *User) bool {
	return p.IsOwnedBy(u) || (u.IsAdmin() && p.WasSubmitted())
}

func (p *Proposal) IsAmendable() bool {
	return p.Status == ProposalStatusDraft ||
		p.Status == ProposalStatusChangesRequested
}
