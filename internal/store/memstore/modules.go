package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type ModuleStore struct {
	mu              sync.RWMutex
	modulesByID     map[int64]domain.Module
	modulesByCourse map[int64][]int64
	nextID          int64
}

func NewModuleStore() *ModuleStore {
	return &ModuleStore{
		modulesByID:     make(map[int64]domain.Module),
		modulesByCourse: make(map[int64][]int64),
		nextID:          1,
	}
}

func (s *ModuleStore) CreateModule(ctx context.Context, m *domain.Module) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m.ID = s.nextID
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now

	courseModules := s.modulesByCourse[m.CourseID]
	m.Position = len(courseModules) + 1

	s.modulesByID[m.ID] = *m
	s.modulesByCourse[m.CourseID] = append(courseModules, m.ID)
	s.nextID++

	return nil
}

func (s *ModuleStore) GetModuleByID(ctx context.Context, id int64) (*domain.Module, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if m, ok := s.modulesByID[id]; ok {
		copy := m
		return &copy, true
	}
	return nil, false
}

func (s *ModuleStore) UpdateModule(ctx context.Context, m *domain.Module) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.modulesByID[m.ID]; !exists {
		return errors.New("module does not exist")
	}

	m.UpdatedAt = time.Now()
	s.modulesByID[m.ID] = *m
	return nil
}

func (s *ModuleStore) DeleteModuleByID(ctx context.Context, id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, exists := s.modulesByID[id]
	if !exists {
		return errors.New("module does not exist")
	}

	courseModules := s.modulesByCourse[m.CourseID]
	for i, mid := range courseModules {
		if mid == id {
			s.modulesByCourse[m.CourseID] = append(courseModules[:i], courseModules[i+1:]...)
			break
		}
	}

	delete(s.modulesByID, id)
	return nil
}

func (s *ModuleStore) ListModulesByCourseID(ctx context.Context, courseID int64) ([]domain.Module, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	moduleIDs := s.modulesByCourse[courseID]
	out := make([]domain.Module, 0, len(moduleIDs))

	for _, id := range moduleIDs {
		if m, ok := s.modulesByID[id]; ok {
			out = append(out, m)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Position < out[j].Position
	})

	return out, nil
}

func (s *ModuleStore) ReorderModules(ctx context.Context, courseID int64, moduleIDs []int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for i, id := range moduleIDs {
		m, exists := s.modulesByID[id]
		if !exists {
			return errors.New("module does not exist")
		}
		if m.CourseID != courseID {
			return errors.New("module does not belong to course")
		}
		m.Position = i + 1
		m.UpdatedAt = now
		s.modulesByID[id] = m
	}

	s.modulesByCourse[courseID] = moduleIDs

	return nil
}
