package memory

import (
	"context"
	"sync"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
)

var (
	_ persistence.EnrollmentRepository = (*EnrollmentRepository)(nil)
)

type EnrollmentRepository struct {
	mu          sync.RWMutex
	enrollments map[int64]map[int64]domain.Enrollment
}

func NewEnrollmentRepository() *EnrollmentRepository {
	return &EnrollmentRepository{
		enrollments: make(map[int64]map[int64]domain.Enrollment),
	}
}

func (r *EnrollmentRepository) Create(ctx context.Context, e *domain.Enrollment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.enrollments[e.UserID] == nil {
		r.enrollments[e.UserID] = make(map[int64]domain.Enrollment)
	}

	if _, exists := r.enrollments[e.UserID][e.CourseID]; exists {
		return errors.ErrConflict
	}

	e.EnrolledAt = time.Now()
	r.enrollments[e.UserID][e.CourseID] = *e

	return nil
}

func (r *EnrollmentRepository) GetByUserAndCourse(ctx context.Context, userID, courseID int64) (*domain.Enrollment, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userEnrollments, ok := r.enrollments[userID]
	if !ok {
		return nil, false
	}

	enrollment, ok := userEnrollments[courseID]
	if !ok {
		return nil, false
	}

	return &enrollment, true
}

func (r *EnrollmentRepository) ListByUser(ctx context.Context, userID int64) ([]domain.Enrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userEnrollments, ok := r.enrollments[userID]
	if !ok {
		return []domain.Enrollment{}, nil
	}

	result := make([]domain.Enrollment, 0, len(userEnrollments))
	for _, e := range userEnrollments {
		result = append(result, e)
	}

	return result, nil
}

func (r *EnrollmentRepository) ListByCourse(ctx context.Context, courseID int64) ([]domain.Enrollment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Enrollment, 0)
	for _, userEnrollments := range r.enrollments {
		if enrollment, ok := userEnrollments[courseID]; ok {
			result = append(result, enrollment)
		}
	}

	return result, nil
}

func (r *EnrollmentRepository) Delete(ctx context.Context, userID, courseID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userEnrollments, ok := r.enrollments[userID]
	if !ok {
		return errors.ErrNotFound
	}

	if _, ok := userEnrollments[courseID]; !ok {
		return errors.ErrNotFound
	}

	delete(userEnrollments, courseID)
	if len(userEnrollments) == 0 {
		delete(r.enrollments, userID)
	}

	return nil
}
