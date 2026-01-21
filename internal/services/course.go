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

var (
	_ Message = (*CreateCourseCommand)(nil)
	_ Message = (*UpdateCourseCommand)(nil)
	_ Message = (*PublishCourseCommand)(nil)
	_ Message = (*CreateCourseFromProposalCommand)(nil)
	_ Message = (*GetCourseByIDQuery)(nil)
)

type CourseService struct {
	Courses   persistence.CourseRepository
	Proposals persistence.ProposalRepository
	Events    events.EventBus
}

func NewCourseService(
	courses persistence.CourseRepository,
	proposals persistence.ProposalRepository,
	eventBus events.EventBus,
) *CourseService {
	return &CourseService{
		Courses:   courses,
		Proposals: proposals,
		Events:    eventBus,
	}
}

// CreateCourseCommand contains the data needed to create a course.
type CreateCourseCommand struct {
	InstructorID         int64
	Title                string
	Summary              string
	TargetAudience       string
	LearningObjectives   string
	AssumedPrerequisites string
}

func (i *CreateCourseCommand) Validate(v *validation.Validator) {
	v.Field(i.InstructorID, "instructor_id").EntityID()
	v.Field(i.Title, "title").Required().MinLength(4).MaxLength(128)
	v.Field(i.Summary, "summary").Required().MaxLength(2048)
}

// Create creates a new course.
func (s *CourseService) Create(ctx context.Context, input *CreateCourseCommand) (*domain.Course, error) {
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

	if err := s.Courses.Create(ctx, course); err != nil {
		return nil, err
	}

	return course, nil
}

// UpdateCourseCommand contains the data needed to update a course.
type UpdateCourseCommand struct {
	CourseID             int64
	UserID               int64
	Title                string
	Summary              string
	TargetAudience       string
	LearningObjectives   string
	AssumedPrerequisites string
}

func (i *UpdateCourseCommand) Validate(v *validation.Validator) {
	v.Field(i.CourseID, "course_id").EntityID()
	v.Field(i.UserID, "user_id").EntityID()
	v.Field(i.Title, "title").Required().MinLength(4).MaxLength(128)
	v.Field(i.Summary, "summary").Required().MaxLength(2048)
}

// Update updates an existing course.
func (s *CourseService) Update(ctx context.Context, input *UpdateCourseCommand) (*domain.Course, error) {
	if err := validation.New().Validate(input); err != nil {
		return nil, err
	}

	course, ok := s.Courses.GetByID(ctx, input.CourseID)
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

	if err := s.Courses.Update(ctx, course); err != nil {
		return nil, err
	}

	event := domain.NewCourseUpdatedEvent(course.ID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return course, nil
}

// PublishCourseCommand contains the data needed to publish a course.
type PublishCourseCommand struct {
	CourseID int64
	UserID   int64
}

func (c *PublishCourseCommand) Validate(v *validation.Validator) {
	v.Field(c.CourseID, "course_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

// Publish publishes a draft course.
func (s *CourseService) Publish(ctx context.Context, input *PublishCourseCommand) (*domain.Course, error) {
	course, ok := s.Courses.GetByID(ctx, input.CourseID)
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

	if err := s.Courses.Update(ctx, course); err != nil {
		return nil, err
	}

	event := domain.NewCoursePublishedEvent(course.ID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return course, nil
}

// CreateCourseFromProposalCommand contains the data needed to create a course from a proposal.
type CreateCourseFromProposalCommand struct {
	ProposalID int64
	UserID     int64
}

func (c *CreateCourseFromProposalCommand) Validate(v *validation.Validator) {
	v.Field(c.ProposalID, "proposal_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

// CreateFromProposal creates a course from an approved proposal.
func (s *CourseService) CreateFromProposal(ctx context.Context, input *CreateCourseFromProposalCommand) (*domain.Course, error) {
	proposal, ok := s.Proposals.GetByID(ctx, input.ProposalID)
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
	if _, ok := s.Courses.GetByProposalID(ctx, proposal.ID); ok {
		return nil, errors.ErrConflict
	}

	course := domain.CourseFromProposal(proposal)

	if err := s.Courses.Create(ctx, course); err != nil {
		return nil, err
	}

	event := domain.NewCourseCreatedFromProposalEvent(course.ID, proposal.ID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return course, nil
}

// GetCourseByIDQuery contains the data needed to get a course by ID.
type GetCourseByIDQuery struct {
	CourseID int64
	UserID   int64
	IsAdmin  bool
}

func (q *GetCourseByIDQuery) Validate(v *validation.Validator) {
	v.Field(q.CourseID, "course_id").EntityID()
	v.Field(q.UserID, "user_id").EntityID()
}

// GetByID retrieves a course by ID with access control.
func (s *CourseService) GetByID(ctx context.Context, input *GetCourseByIDQuery) (*domain.Course, error) {
	course, ok := s.Courses.GetByID(ctx, input.CourseID)
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
	return s.Courses.ListAllLive(ctx)
}
