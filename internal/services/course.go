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

// CourseService handles all course operations.
type CourseService struct {
	courses   persistence.CourseRepository
	proposals persistence.ProposalRepository
	events    events.EventBus
}

// NewCourseService creates a new CourseService with the given dependencies.
func NewCourseService(
	courses persistence.CourseRepository,
	proposals persistence.ProposalRepository,
	eventBus events.EventBus,
) *CourseService {
	return &CourseService{
		courses:   courses,
		proposals: proposals,
		events:    eventBus,
	}
}

// CreateCourseInput contains the data needed to create a course.
type CreateCourseInput struct {
	InstructorID         int64
	Title                string
	Summary              string
	TargetAudience       string
	LearningObjectives   string
	AssumedPrerequisites string
}

func (i *CreateCourseInput) Validate(v *validation.Validator) {
	v.Field(i.InstructorID, "instructor_id").EntityID()
	v.Field(i.Title, "title").Required().MinLength(4).MaxLength(128)
	v.Field(i.Summary, "summary").Required().MaxLength(2048)
}

// Create creates a new course.
func (s *CourseService) Create(ctx context.Context, input *CreateCourseInput) (*domain.Course, error) {
	if err := validation.New().Validate(input); err != nil {
		return nil, err
	}

	course := &domain.Course{
		InstructorID:         input.InstructorID,
		Title:                strings.TrimSpace(input.Title),
		Summary:              strings.TrimSpace(input.Summary),
		TargetAudience:       strings.TrimSpace(input.TargetAudience),
		LearningObjectives:   strings.TrimSpace(input.LearningObjectives),
		AssumedPrerequisites: strings.TrimSpace(input.AssumedPrerequisites),
		Status:               domain.CourseStatusDraft,
	}

	if err := s.courses.Create(ctx, course); err != nil {
		return nil, err
	}

	return course, nil
}

// UpdateCourseInput contains the data needed to update a course.
type UpdateCourseInput struct {
	CourseID             int64
	UserID               int64
	Title                string
	Summary              string
	TargetAudience       string
	LearningObjectives   string
	AssumedPrerequisites string
}

func (i *UpdateCourseInput) Validate(v *validation.Validator) {
	v.Field(i.CourseID, "course_id").EntityID()
	v.Field(i.UserID, "user_id").EntityID()
	v.Field(i.Title, "title").Required().MinLength(4).MaxLength(128)
	v.Field(i.Summary, "summary").Required().MaxLength(2048)
}

// Update updates an existing course.
func (s *CourseService) Update(ctx context.Context, input *UpdateCourseInput) (*domain.Course, error) {
	if err := validation.New().Validate(input); err != nil {
		return nil, err
	}

	course, ok := s.courses.GetByID(ctx, input.CourseID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if course.InstructorID != input.UserID {
		return nil, errors.ErrForbidden
	}

	if !course.IsAmendable() {
		return nil, errors.ErrInvalidStatusTransition
	}

	course.Title = strings.TrimSpace(input.Title)
	course.Summary = strings.TrimSpace(input.Summary)
	course.TargetAudience = strings.TrimSpace(input.TargetAudience)
	course.LearningObjectives = strings.TrimSpace(input.LearningObjectives)
	course.AssumedPrerequisites = strings.TrimSpace(input.AssumedPrerequisites)

	if err := s.courses.Update(ctx, course); err != nil {
		return nil, err
	}

	event := domain.NewCourseUpdatedEvent(course.ID, course.InstructorID)
	_ = s.events.Publish(ctx, event)

	return course, nil
}

// PublishCourseInput contains the data needed to publish a course.
type PublishCourseInput struct {
	CourseID int64
	UserID   int64
}

// Publish publishes a draft course.
func (s *CourseService) Publish(ctx context.Context, input *PublishCourseInput) (*domain.Course, error) {
	course, ok := s.courses.GetByID(ctx, input.CourseID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if course.InstructorID != input.UserID {
		return nil, errors.ErrForbidden
	}

	if course.Status != domain.CourseStatusDraft {
		return nil, errors.ErrInvalidStatusTransition
	}

	course.Status = domain.CourseStatusLive

	if err := s.courses.Update(ctx, course); err != nil {
		return nil, err
	}

	event := domain.NewCoursePublishedEvent(course.ID, course.InstructorID)
	_ = s.events.Publish(ctx, event)

	return course, nil
}

// CreateFromProposalInput contains the data needed to create a course from a proposal.
type CreateFromProposalInput struct {
	ProposalID int64
	UserID     int64
}

// CreateFromProposal creates a course from an approved proposal.
func (s *CourseService) CreateFromProposal(ctx context.Context, input *CreateFromProposalInput) (*domain.Course, error) {
	proposal, ok := s.proposals.GetByID(ctx, input.ProposalID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	if proposal.AuthorID != input.UserID {
		return nil, errors.ErrForbidden
	}

	if proposal.Status != domain.ProposalStatusApproved {
		return nil, errors.ErrInvalidStatusTransition
	}

	// Check if course already exists for this proposal
	if _, ok := s.courses.GetByProposalID(ctx, proposal.ID); ok {
		return nil, errors.ErrConflict
	}

	course := domain.CourseFromProposal(proposal)

	if err := s.courses.Create(ctx, course); err != nil {
		return nil, err
	}

	event := domain.NewCourseCreatedFromProposalEvent(course.ID, proposal.ID, course.InstructorID)
	_ = s.events.Publish(ctx, event)

	return course, nil
}

// GetByIDInput contains the data needed to get a course by ID.
type GetCourseByIDInput struct {
	CourseID int64
	UserID   int64
	IsAdmin  bool
}

// GetByID retrieves a course by ID with access control.
func (s *CourseService) GetByID(ctx context.Context, input *GetCourseByIDInput) (*domain.Course, error) {
	course, ok := s.courses.GetByID(ctx, input.CourseID)
	if !ok {
		return nil, errors.ErrNotFound
	}

	// Check visibility
	isOwner := course.InstructorID == input.UserID
	if !course.IsLive() && !isOwner && !input.IsAdmin {
		return nil, errors.ErrForbidden
	}

	return course, nil
}

// ListLive retrieves all live courses.
func (s *CourseService) ListLive(ctx context.Context) ([]domain.Course, error) {
	return s.courses.ListAllLive(ctx)
}
