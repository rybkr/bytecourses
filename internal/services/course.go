package services

import (
	"context"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/pkg/events"
	"bytecourses/internal/pkg/validation"
)

var (
	_ Command = (*CreateCourseCommand)(nil)
	_ Command = (*UpdateCourseCommand)(nil)
	_ Command = (*PublishCourseCommand)(nil)
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

type CreateCourseCommand struct {
	InstructorID         int64  `json:"instructor_id"`
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (c *CreateCourseCommand) Validate(v *validation.Validator) {
	v.Field(c.InstructorID, "instructor_id").EntityID()
	v.Field(c.Title, "title").Required().MinLength(4).MaxLength(128).IsTrimmed()
	v.Field(c.Summary, "summary").Required().MaxLength(2048).IsTrimmed()
	v.Field(c.TargetAudience, "target_audience").Required().MaxLength(2048).IsTrimmed()
	v.Field(c.LearningObjectives, "learning_objectives").Required().MaxLength(2048).IsTrimmed()
	v.Field(c.AssumedPrerequisites, "assumed_prerequisites").Required().MaxLength(2048).IsTrimmed()
}

func (s *CourseService) Create(ctx context.Context, cmd *CreateCourseCommand) (*domain.Course, error) {
	if err := validation.Validate(cmd); err != nil {
		return nil, err
	}

	course := domain.Course{
		InstructorID:         cmd.InstructorID,
		Title:                cmd.Title,
		Summary:              cmd.Summary,
		TargetAudience:       cmd.TargetAudience,
		LearningObjectives:   cmd.LearningObjectives,
		AssumedPrerequisites: cmd.AssumedPrerequisites,
		Status:               domain.CourseStatusDraft,
	}
	if err := s.Courses.Create(ctx, &course); err != nil {
		return nil, err
	}

	event := domain.NewCourseCreatedEvent(course.ID, cmd.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return &course, nil
}

type UpdateCourseCommand struct {
	CourseID             int64  `json:"course_id"`
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
	UserID               int64  `json:"user_id"`
}

func (c *UpdateCourseCommand) Validate(v *validation.Validator) {
	v.Field(c.CourseID, "instructor_id").EntityID()
	v.Field(c.Title, "title").Required().MinLength(4).MaxLength(128).IsTrimmed()
	v.Field(c.Summary, "summary").Required().MaxLength(2048).IsTrimmed()
	v.Field(c.TargetAudience, "target_audience").Required().MaxLength(2048).IsTrimmed()
	v.Field(c.LearningObjectives, "learning_objectives").Required().MaxLength(2048).IsTrimmed()
	v.Field(c.AssumedPrerequisites, "assumed_prerequisites").Required().MaxLength(2048).IsTrimmed()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *CourseService) Update(ctx context.Context, cmd *UpdateCourseCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	course, ok := s.Courses.GetByID(ctx, cmd.CourseID)
	if !ok {
		return errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return errors.ErrNotFound
	}

	course.Title = cmd.Title
	course.Summary = cmd.Summary
	course.TargetAudience = cmd.TargetAudience
	course.LearningObjectives = cmd.LearningObjectives
	course.AssumedPrerequisites = cmd.AssumedPrerequisites
	if err := s.Courses.Update(ctx, course); err != nil {
		return err
	}

	event := domain.NewCourseUpdatedEvent(course.ID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type PublishCourseCommand struct {
    CourseID int64`json:"course_id"`
    UserID   int64`json:"user_id"`
}

func (c *PublishCourseCommand) Validate(v *validation.Validator) {
	v.Field(c.CourseID, "course_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *CourseService) Publish(ctx context.Context, cmd *PublishCourseCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	course, ok := s.Courses.GetByID(ctx, cmd.CourseID)
	if !ok {
		return errors.ErrNotFound
	}
	if course.InstructorID != cmd.UserID {
		return errors.ErrNotFound
	}
	if course.Status != domain.CourseStatusDraft {
		return errors.ErrInvalidStatusTransition
	}

	course.Status = domain.CourseStatusPublished
	if err := s.Courses.Update(ctx, course); err != nil {
		return err
	}

	event := domain.NewCoursePublishedEvent(course.ID, course.InstructorID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type GetCourseQuery struct {
    CourseID int64`json:"course_id"`
    UserID   int64`json:"user_id"`
    UserRole domain.UserRole`json:"user_role"`
}

func (s *CourseService) Get(ctx context.Context, query *GetCourseQuery) (*domain.Course, error) {
	course, ok := s.Courses.GetByID(ctx, query.CourseID)
	if !ok {
		return nil, errors.ErrNotFound
	}

    switch query.UserRole {
    case domain.UserRoleStudent,
        domain.UserRoleInstructor:
        if course.InstructorID != query.UserID {
            return nil, errors.ErrNotFound
        }

    case domain.UserRoleAdmin:

    default:
        return nil, errors.ErrForbidden
    }

	return course, nil
}

func (s *CourseService) List(ctx context.Context) ([]domain.Course, error) {
	return s.Courses.ListAllLive(ctx)
}
