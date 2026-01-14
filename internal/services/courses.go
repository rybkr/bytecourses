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

func (s *CourseService) CreateCourseFromProposal(ctx context.Context, proposal *domain.Proposal, user *domain.User) (*domain.Course, error) {
	if proposal.Status != domain.ProposalStatusApproved {
		return nil, ErrInvalidInput
	}
	if proposal.AuthorID != user.ID {
		return nil, ErrForbidden
	}

	existing, _ := s.courses.GetCourseByProposalID(ctx, proposal.ID)
	if existing != nil {
		return nil, ErrConflict
	}

	course := domain.CourseFromProposal(proposal)
	if err := s.courses.CreateCourse(ctx, course); err != nil {
		s.logger.Error("course creation from proposal failed",
			"event", "course.creation.from_proposal",
			"proposal_id", proposal.ID,
			"user_id", user.ID,
			"title", proposal.Title,
			"error", err,
		)
		return nil, err
	}

	s.logger.Info("course.created.from_proposal",
		"course_id", course.ID,
		"proposal_id", proposal.ID,
		"user_id", user.ID,
		"title", course.Title,
		"status", course.Status,
	)

	return course, nil
}
