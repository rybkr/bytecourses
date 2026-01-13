package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"sync"
	"time"
)

type CourseStore struct {
	mu          sync.RWMutex
	coursesByID map[int64]domain.Course
	nextID      int64
}

func NewCourseStore() *CourseStore {
	return &CourseStore{
		coursesByID: make(map[int64]domain.Course),
		nextID:      1,
	}
}

func (s *CourseStore) CreateCourse(ctx context.Context, c *domain.Course) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c.ID = s.nextID
	c.CreatedAt = time.Now()

	s.coursesByID[c.ID] = *c
	s.nextID++

	return nil
}

func (s *CourseStore) GetCourseByID(ctx context.Context, id int64) (*domain.Course, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if c, ok := s.coursesByID[id]; ok {
		return &c, true
	}
	return nil, false
}

func (s *CourseStore) ListAllLiveCourses(ctx context.Context) ([]domain.Course, error) {
	out := make([]domain.Course, 0)
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, c := range s.coursesByID {
		if c.IsLive() {
			out = append(out, c)
		}
	}

	return out, nil
}
