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
	_ Command = (*EnrollCommand)(nil)
	_ Command = (*UnenrollCommand)(nil)
)

type EnrollmentService struct {
	Enrollments persistence.EnrollmentRepository
	Courses     persistence.CourseRepository
	Users       persistence.UserRepository
	Events      events.EventBus
}

func NewEnrollmentService(
	enrollments persistence.EnrollmentRepository,
	courses persistence.CourseRepository,
	users persistence.UserRepository,
	eventBus events.EventBus,
) *EnrollmentService {
	return &EnrollmentService{
		Enrollments: enrollments,
		Courses:     courses,
		Users:       users,
		Events:      eventBus,
	}
}

type EnrollCommand struct {
	CourseID int64 `json:"course_id"`
	UserID   int64 `json:"user_id"`
}

func (c *EnrollCommand) Validate(v *validation.Validator) {
	v.Field(c.CourseID, "course_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *EnrollmentService) Enroll(ctx context.Context, cmd *EnrollCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	course, ok := s.Courses.GetByID(ctx, cmd.CourseID)
	if !ok {
		return errors.ErrNotFound
	}

	if !course.IsLive() {
		return errors.ErrInvalidStatusTransition
	}

	_, ok = s.Users.GetByID(ctx, cmd.UserID)
	if !ok {
		return errors.ErrNotFound
	}

	_, ok = s.Enrollments.GetByUserAndCourse(ctx, cmd.UserID, cmd.CourseID)
	if ok {
		return errors.ErrConflict
	}

	enrollment := &domain.Enrollment{
		UserID:   cmd.UserID,
		CourseID: cmd.CourseID,
	}

	if err := s.Enrollments.Create(ctx, enrollment); err != nil {
		return err
	}

	event := domain.NewEnrollmentCreatedEvent(cmd.UserID, cmd.CourseID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type UnenrollCommand struct {
	CourseID int64 `json:"course_id"`
	UserID   int64 `json:"user_id"`
}

func (c *UnenrollCommand) Validate(v *validation.Validator) {
	v.Field(c.CourseID, "course_id").EntityID()
	v.Field(c.UserID, "user_id").EntityID()
}

func (s *EnrollmentService) Unenroll(ctx context.Context, cmd *UnenrollCommand) error {
	if err := validation.Validate(cmd); err != nil {
		return err
	}

	_, ok := s.Enrollments.GetByUserAndCourse(ctx, cmd.UserID, cmd.CourseID)
	if !ok {
		return errors.ErrNotFound
	}

	if err := s.Enrollments.Delete(ctx, cmd.UserID, cmd.CourseID); err != nil {
		return err
	}

	event := domain.NewEnrollmentDeletedEvent(cmd.UserID, cmd.CourseID)
	_ = s.Events.Publish(ctx, event)

	return nil
}

type IsEnrolledQuery struct {
	CourseID int64 `json:"course_id"`
	UserID   int64 `json:"user_id"`
}

func (s *EnrollmentService) IsEnrolled(ctx context.Context, query *IsEnrolledQuery) (bool, error) {
	_, ok := s.Enrollments.GetByUserAndCourse(ctx, query.UserID, query.CourseID)
	return ok, nil
}

type ListEnrollmentsByUserQuery struct {
	UserID int64 `json:"user_id"`
}

func (s *EnrollmentService) ListByUser(ctx context.Context, query *ListEnrollmentsByUserQuery) ([]domain.Enrollment, error) {
	return s.Enrollments.ListByUser(ctx, query.UserID)
}

type ListEnrollmentsByCourseQuery struct {
	CourseID int64 `json:"course_id"`
}

func (s *EnrollmentService) ListByCourse(ctx context.Context, query *ListEnrollmentsByCourseQuery) ([]domain.Enrollment, error) {
	return s.Enrollments.ListByCourse(ctx, query.CourseID)
}
