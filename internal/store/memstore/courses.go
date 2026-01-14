package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"errors"
	"sync"
	"time"
)

type CourseStore struct {
	mu                sync.RWMutex
	coursesByID       map[int64]domain.Course
	coursesByProposal map[int64]int64
	nextID            int64
}

func NewCourseStore() *CourseStore {
	return &CourseStore{
		coursesByID:       make(map[int64]domain.Course),
		coursesByProposal: make(map[int64]int64),
		nextID:            1,
	}
}

func (s *CourseStore) CreateCourse(ctx context.Context, c *domain.Course) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c.ID = s.nextID
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now

	s.coursesByID[c.ID] = *c
	if c.ProposalID != nil {
		s.coursesByProposal[*c.ProposalID] = c.ID
	}
	s.nextID++

	return nil
}

func (s *CourseStore) GetCourseByID(ctx context.Context, id int64) (*domain.Course, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if c, ok := s.coursesByID[id]; ok {
		copy := c
		return &copy, true
	}
	return nil, false
}

func (s *CourseStore) GetCourseByProposalID(ctx context.Context, proposalID int64) (*domain.Course, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	courseID, ok := s.coursesByProposal[proposalID]
	if !ok {
		return nil, false
	}

	c, ok := s.coursesByID[courseID]
	if !ok {
		return nil, false
	}

	copy := c
	return &copy, true
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

func (s *CourseStore) UpdateCourse(ctx context.Context, c *domain.Course) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.coursesByID[c.ID]; !exists {
		return errors.New("course does not exist")
	}

	c.UpdatedAt = time.Now()
	s.coursesByID[c.ID] = *c
	return nil
}
