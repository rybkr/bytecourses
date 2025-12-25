package domain

import (
	"time"
)

// UserRole defines authorization levels for a user.
type UserRole string

const (
	UserRoleStudent    UserRole = "student"
	UserRoleInstructor UserRole = "instructor"
	UserRoleAdmin      UserRole = "admin"
)

// ProposalStatus is the lifecycle state of a course proposal.
type ProposalStatus string

const (
	ProposalSubmitted        ProposalStatus = "submitted"
	ProposalChangesRequested ProposalStatus = "changes_requested"
	ProposalApproved         ProposalStatus = "approved"
	ProposalDenied           ProposalStatus = "denied"
)

// User represents an authenticated actor in the system.
type User struct {
	ID           int64
	Email        string
	PasswordHash []byte
	Role         UserRole
	CreatedAt    time.Time
}

// CourseProposal represents an instructor application.
type CourseProposal struct {
	ID           int64
	InstructorID int64
	Title        string
	Summary      string
	Status       ProposalStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
