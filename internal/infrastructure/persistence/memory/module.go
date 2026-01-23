package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

var (
	_ persistence.ModuleRepository = (*ModuleRepository)(nil)
)

type ModuleRepository struct {
	mu      sync.RWMutex
	modules map[int64]domain.Module
	nextID  int64
}

func NewModuleRepository() *ModuleRepository {
	return &ModuleRepository{
		modules: make(map[int64]domain.Module),
		nextID:  1,
	}
}

func (r *ModuleRepository) Create(ctx context.Context, m *domain.Module) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m.ID = r.nextID
	r.nextID++
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()

	r.modules[m.ID] = *m
	return nil
}

func (r *ModuleRepository) GetByID(ctx context.Context, id int64) (*domain.Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.modules[id]
	if !ok {
		return nil, false
	}

	return &m, true
}

func (r *ModuleRepository) Update(ctx context.Context, m *domain.Module) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.modules[m.ID]; !ok {
		return nil
	}

	m.UpdatedAt = time.Now()
	r.modules[m.ID] = *m
	return nil
}

func (r *ModuleRepository) ListByCourseID(ctx context.Context, courseID int64) ([]domain.Module, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Module, 0)
	for _, m := range r.modules {
		if m.CourseID == courseID {
			result = append(result, m)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})

	return result, nil
}

func (r *ModuleRepository) DeleteByID(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.modules, id)
	return nil
}
