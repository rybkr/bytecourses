package services

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"log/slog"
)

type CourseService struct {
	courses store.CourseStore
	logger  *CourseLogger
}

func NewCourseService(courses store.CourseStore, logger *slog.Logger) *CourseService {
	return &CourseService{
		courses: courses,
		logger:  NewCourseLogger(logger),
	}
}

type CreateCourseRequest struct {
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	InstructorID int64  `json:"instructor_id"`
}

func (r *CreateCourseRequest) IsValid() bool {
	return r.InstructorID > 0 && r.Title != ""
}

func (s *CourseService) CreateCourse(ctx context.Context, request *CreateCourseRequest) (*domain.Course, error) {
	if !request.IsValid() {
		return nil, ErrInvalidInput
	}

	course := &domain.Course{
		InstructorID: request.InstructorID,
		Title:        request.Title,
		Summary:      request.Summary,
		Status:       domain.CourseStatusDraft,
	}
	if err := s.courses.CreateCourse(ctx, course); err != nil {
		s.logger.Error("course creation failed",
			"event", "course.creation",
			"user_id", request.InstructorID,
			"title", request.Title,
			"error", err,
		)
		return nil, err
	}

	s.logger.Info("course.created",
		"course_id", course.ID,
		"user_id", request.InstructorID,
		"title", course.Title,
		"status", course.Status,
	)

	return course, nil
}

func (s *CourseService) GetCourse(ctx context.Context, c *domain.Course, u *domain.User) (*domain.Course, error) {
	if !c.IsViewableBy(u) {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *CourseService) ListCourses(ctx context.Context) ([]domain.Course, error) {
	return s.courses.ListAllLiveCourses(ctx)
}
