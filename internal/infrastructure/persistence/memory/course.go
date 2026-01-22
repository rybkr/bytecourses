package memory

import (
	"context"
	"sync"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

var (
	_ persistence.CourseRepository = (*CourseRepository)(nil)
)

type CourseRepository struct {
	mu      sync.RWMutex
	courses map[int64]domain.Course
	nextID  int64
}

func NewCourseRepository() *CourseRepository {
	return &CourseRepository{
		courses: make(map[int64]domain.Course),
		nextID:  1,
	}
}

func (r *CourseRepository) Create(ctx context.Context, c *domain.Course) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c.ID = r.nextID
	r.nextID++
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	r.courses[c.ID] = *c

	return nil
}

func (r *CourseRepository) GetByID(ctx context.Context, id int64) (*domain.Course, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.courses[id]
	if !ok {
		return nil, false
	}

	return &c, true
}

func (r *CourseRepository) ListAllLive(ctx context.Context) ([]domain.Course, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Course, 0)
	for _, c := range r.courses {
		if c.Status == domain.CourseStatusPublished {
			result = append(result, c)
		}
	}
	return result, nil
}

func (r *CourseRepository) Update(ctx context.Context, c *domain.Course) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.courses[c.ID]; !ok {
		return nil
	}

	c.UpdatedAt = time.Now()
	r.courses[c.ID] = *c

	return nil
}
