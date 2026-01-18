package domain

import (
	"time"
)

type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

type BaseEvent struct {
	occurredAt time.Time
}

func NewBaseEvent() BaseEvent {
	return BaseEvent{
		occurredAt: time.Now(),
	}
}

func (e *BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

type UserRegisteredEvent struct {
	BaseEvent
	UserID int64
	Email  string
	Name   string
}

func NewUserRegisteredEvent(userID int64, email string, name string) *UserRegisteredEvent {
	return &UserRegisteredEvent{
		BaseEvent: NewBaseEvent(),
		UserID:    userID,
		Email:     email,
		Name:      name,
	}
}

func (e *UserRegisteredEvent) EventName() string {
	return "user.registered"
}

type UserProfileUpdatedEvent struct {
	BaseEvent
	UserID  int64
	OldName string
	NewName string
}

func NewUserProfileUpdatedEvent(userID int64, oldName, newName string) *UserProfileUpdatedEvent {
	return &UserProfileUpdatedEvent{
		BaseEvent: NewBaseEvent(),
		UserID:    userID,
		OldName:   oldName,
		NewName:   newName,
	}
}

func (e *UserProfileUpdatedEvent) EventName() string {
	return "user.profile_updated"
}

type PasswordResetRequestedEvent struct {
	BaseEvent
	UserID int64
	Email  string
}

func NewPasswordResetRequestedEvent(userID int64, email string) *PasswordResetRequestedEvent {
	return &PasswordResetRequestedEvent{
		BaseEvent: NewBaseEvent(),
		UserID:    userID,
		Email:     email,
	}
}

func (e *PasswordResetRequestedEvent) EventName() string {
	return "user.password_reset_requested"
}

type PasswordResetCompletedEvent struct {
	BaseEvent
	UserID int64
}

func NewPasswordResetCompletedEvent(userID int64) *PasswordResetCompletedEvent {
	return &PasswordResetCompletedEvent{
		BaseEvent: NewBaseEvent(),
		UserID:    userID,
	}
}

func (e *PasswordResetCompletedEvent) EventName() string {
	return "user.password_reset_completed"
}

type ProposalCreatedEvent struct {
	BaseEvent
	ProposalID int64
	AuthorID   int64
	Title      string
}

func NewProposalCreatedEvent(proposalID int64, authorID int64, title string) *ProposalCreatedEvent {
	return &ProposalCreatedEvent{
		BaseEvent:  NewBaseEvent(),
		ProposalID: proposalID,
		AuthorID:   authorID,
		Title:      title,
	}
}

func (e *ProposalCreatedEvent) EventName() string {
	return "proposal.created"
}

type ProposalUpdatedEvent struct {
	BaseEvent
	ProposalID int64
	AuthorID   int64
	Title      string
}

func NewProposalUpdatedEvent(proposalID int64, authorID int64, title string) *ProposalUpdatedEvent {
	return &ProposalUpdatedEvent{
		BaseEvent:  NewBaseEvent(),
		ProposalID: proposalID,
		AuthorID:   authorID,
		Title:      title,
	}
}

func (e *ProposalUpdatedEvent) EventName() string {
	return "proposal.updated"
}

type ProposalSubmittedEvent struct {
	BaseEvent
	ProposalID int64
	AuthorID   int64
	Title      string
	OldStatus  ProposalStatus
}

func NewProposalSubmittedEvent(proposalID int64, authorID int64, title string, oldStatus ProposalStatus) *ProposalSubmittedEvent {
	return &ProposalSubmittedEvent{
		BaseEvent:  NewBaseEvent(),
		ProposalID: proposalID,
		AuthorID:   authorID,
		Title:      title,
		OldStatus:  oldStatus,
	}
}

func (e *ProposalSubmittedEvent) EventName() string {
	return "proposal.submitted"
}

type ProposalWithdrawnEvent struct {
	BaseEvent
	ProposalID int64
	AuthorID   int64
}

func NewProposalWithdrawnEvent(proposalID int64, authorID int64) *ProposalWithdrawnEvent {
	return &ProposalWithdrawnEvent{
		BaseEvent:  NewBaseEvent(),
		ProposalID: proposalID,
		AuthorID:   authorID,
	}
}

func (e *ProposalWithdrawnEvent) EventName() string {
	return "proposal.withdrawn"
}

type ProposalApprovedEvent struct {
	BaseEvent
	ProposalID int64
	AuthorID   int64
	ReviewerID int64
	Title      string
}

func NewProposalApprovedEvent(proposalID int64, authorID int64, reviewerID int64, title string) *ProposalApprovedEvent {
	return &ProposalApprovedEvent{
		BaseEvent:  NewBaseEvent(),
		ProposalID: proposalID,
		AuthorID:   authorID,
		ReviewerID: reviewerID,
		Title:      title,
	}
}

func (e *ProposalApprovedEvent) EventName() string {
	return "proposal.approved"
}

type ProposalRejectedEvent struct {
	BaseEvent
	ProposalID int64
	AuthorID   int64
	ReviewerID int64
	Title      string
}

func NewProposalRejectedEvent(proposalID int64, authorID int64, reviewerID int64, title string) *ProposalRejectedEvent {
	return &ProposalRejectedEvent{
		BaseEvent:  NewBaseEvent(),
		ProposalID: proposalID,
		AuthorID:   authorID,
		ReviewerID: reviewerID,
		Title:      title,
	}
}

func (e *ProposalRejectedEvent) EventName() string {
	return "proposal.rejected"
}

type ProposalChangesRequestedEvent struct {
	BaseEvent
	ProposalID int64
	AuthorID   int64
	ReviewerID int64
	Title      string
}

func NewProposalChangesRequestedEvent(proposalID int64, authorID int64, reviewerID int64, title string) *ProposalChangesRequestedEvent {
	return &ProposalChangesRequestedEvent{
		BaseEvent:  NewBaseEvent(),
		ProposalID: proposalID,
		AuthorID:   authorID,
		ReviewerID: reviewerID,
		Title:      title,
	}
}

func (e *ProposalChangesRequestedEvent) EventName() string {
	return "proposal.changes_requested"
}

type ProposalDeletedEvent struct {
	BaseEvent
	ProposalID int64
	AuthorID   int64
	Status     ProposalStatus
}

func NewProposalDeletedEvent(proposalID int64, authorID int64, status ProposalStatus) *ProposalDeletedEvent {
	return &ProposalDeletedEvent{
		BaseEvent:  NewBaseEvent(),
		ProposalID: proposalID,
		AuthorID:   authorID,
		Status:     status,
	}
}

func (e *ProposalDeletedEvent) EventName() string {
	return "proposal.deleted"
}

type CourseCreatedFromProposalEvent struct {
	BaseEvent
	CourseID     int64
	ProposalID   int64
	InstructorID int64
	Title        string
}

func NewCourseCreatedFromProposalEvent(courseID int64, proposalID int64, instructorID int64, title string) *CourseCreatedFromProposalEvent {
	return &CourseCreatedFromProposalEvent{
		BaseEvent:    NewBaseEvent(),
		CourseID:     courseID,
		ProposalID:   proposalID,
		InstructorID: instructorID,
		Title:        title,
	}
}

func (e *CourseCreatedFromProposalEvent) EventName() string {
	return "course.created_from_proposal"
}

type CourseUpdatedEvent struct {
	BaseEvent
	CourseID     int64
	InstructorID int64
	Title        string
}

func NewCourseUpdatedEvent(courseID int64, instructorID int64, title string) *CourseUpdatedEvent {
	return &CourseUpdatedEvent{
		BaseEvent:    NewBaseEvent(),
		CourseID:     courseID,
		InstructorID: instructorID,
		Title:        title,
	}
}

func (e *CourseUpdatedEvent) EventName() string {
	return "course.updated"
}

type CoursePublishedEvent struct {
	BaseEvent
	CourseID     int64
	InstructorID int64
	Title        string
}

func NewCoursePublishedEvent(courseID int64, instructorID int64, title string) *CoursePublishedEvent {
	return &CoursePublishedEvent{
		BaseEvent:    NewBaseEvent(),
		CourseID:     courseID,
		InstructorID: instructorID,
		Title:        title,
	}
}

func (e *CoursePublishedEvent) EventName() string {
	return "course.published"
}

type ModuleCreatedEvent struct {
	BaseEvent
	ModuleID     int64
	CourseID     int64
	InstructorID int64
	Title        string
	Position     int
}

func NewModuleCreatedEvent(moduleID int64, courseID int64, instructorID int64, title string, position int) *ModuleCreatedEvent {
	return &ModuleCreatedEvent{
		BaseEvent:    NewBaseEvent(),
		ModuleID:     moduleID,
		CourseID:     courseID,
		InstructorID: instructorID,
		Title:        title,
		Position:     position,
	}
}

func (e *ModuleCreatedEvent) EventName() string {
	return "module.created"
}

type ModuleUpdatedEvent struct {
	BaseEvent
	ModuleID     int64
	CourseID     int64
	InstructorID int64
	Title        string
}

func NewModuleUpdatedEvent(moduleID int64, courseID int64, instructorID int64, title string) *ModuleUpdatedEvent {
	return &ModuleUpdatedEvent{
		BaseEvent:    NewBaseEvent(),
		ModuleID:     moduleID,
		CourseID:     courseID,
		InstructorID: instructorID,
		Title:        title,
	}
}

func (e *ModuleUpdatedEvent) EventName() string {
	return "module.updated"
}

type ModuleDeletedEvent struct {
	BaseEvent
	ModuleID     int64
	CourseID     int64
	InstructorID int64
}

func NewModuleDeletedEvent(moduleID int64, courseID int64, instructorID int64) *ModuleDeletedEvent {
	return &ModuleDeletedEvent{
		BaseEvent:    NewBaseEvent(),
		ModuleID:     moduleID,
		CourseID:     courseID,
		InstructorID: instructorID,
	}
}

func (e *ModuleDeletedEvent) EventName() string {
	return "module.deleted"
}

type ContentCreatedEvent struct {
	BaseEvent
	ContentID    int64
	ModuleID     int64
	CourseID     int64
	InstructorID int64
	Title        string
}

func NewContentCreatedEvent(contentID int64, moduleID int64, courseID int64, instructorID int64, title string) *ContentCreatedEvent {
	return &ContentCreatedEvent{
		BaseEvent:    NewBaseEvent(),
		ContentID:    contentID,
		ModuleID:     moduleID,
		CourseID:     courseID,
		InstructorID: instructorID,
		Title:        title,
	}
}

func (e *ContentCreatedEvent) EventName() string {
	return "content.created"
}

type ContentUpdatedEvent struct {
	BaseEvent
	ContentID    int64
	ModuleID     int64
	CourseID     int64
	InstructorID int64
	Title        string
}

func NewContentUpdatedEvent(contentID int64, moduleID int64, courseID int64, instructorID int64, title string) *ContentUpdatedEvent {
	return &ContentUpdatedEvent{
		BaseEvent:    NewBaseEvent(),
		ContentID:    contentID,
		ModuleID:     moduleID,
		CourseID:     courseID,
		InstructorID: instructorID,
		Title:        title,
	}
}

func (e *ContentUpdatedEvent) EventName() string {
	return "content.updated"
}

type ContentPublishedEvent struct {
	BaseEvent
	ContentID    int64
	ModuleID     int64
	CourseID     int64
	InstructorID int64
	Title        string
}

func NewContentPublishedEvent(contentID int64, moduleID int64, courseID int64, instructorID int64, title string) *ContentPublishedEvent {
	return &ContentPublishedEvent{
		BaseEvent:    NewBaseEvent(),
		ContentID:    contentID,
		ModuleID:     moduleID,
		CourseID:     courseID,
		InstructorID: instructorID,
		Title:        title,
	}
}

func (e *ContentPublishedEvent) EventName() string {
	return "content.published"
}

type ContentDeletedEvent struct {
	BaseEvent
	ContentID    int64
	ModuleID     int64
	CourseID     int64
	InstructorID int64
}

func NewContentDeletedEvent(contentID int64, moduleID int64, courseID int64, instructorID int64) *ContentDeletedEvent {
	return &ContentDeletedEvent{
		BaseEvent:    NewBaseEvent(),
		ContentID:    contentID,
		ModuleID:     moduleID,
		CourseID:     courseID,
		InstructorID: instructorID,
	}
}

func (e *ContentDeletedEvent) EventName() string {
	return "content.deleted"
}
