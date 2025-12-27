package domain

import (
	"time"
)

type UserRole string

const (
	UserRoleStudent    UserRole = "student"
	UserRoleInstructor UserRole = "instructor"
	UserRoleAdmin      UserRole = "admin"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
    Name         string    `json:"name"`
	PasswordHash []byte    `json:"-"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type ProposalStatus string

const (
	ProposalStatusDraft     ProposalStatus = "draft"
	ProposalStatusSubmitted ProposalStatus = "submitted"
	ProposalStatusApproved  ProposalStatus = "approved"
	ProposalStatusRejected  ProposalStatus = "rejected"
)

type Proposal struct {
	ID         int64          `json:"id"`
	Title      string         `json:"title"`
	Summary    string         `json:"summary"`
	AuthorID   int64          `json:"author_id"`
	ReviewerID int64          `json:"reviewer_id"`
	Status     ProposalStatus `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
