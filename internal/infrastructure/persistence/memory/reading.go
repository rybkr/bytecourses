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
	_ persistence.ReadingRepository = (*ReadingRepository)(nil)
)

type ReadingRepository struct {
	mu       sync.RWMutex
	readings map[int64]domain.Reading
	nextID   int64
}

func NewReadingRepository() *ReadingRepository {
	return &ReadingRepository{
		readings: make(map[int64]domain.Reading),
		nextID:   1,
	}
}

func (r *ReadingRepository) Create(ctx context.Context, reading *domain.Reading) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reading.ID = r.nextID
	r.nextID++
	reading.CreatedAt = time.Now()
	reading.UpdatedAt = time.Now()

	r.readings[reading.ID] = *reading
	return nil
}

func (r *ReadingRepository) GetByID(ctx context.Context, id int64) (*domain.Reading, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reading, ok := r.readings[id]
	if !ok {
		return nil, false
	}

	return &reading, true
}

func (r *ReadingRepository) Update(ctx context.Context, reading *domain.Reading) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.readings[reading.ID]; !ok {
		return nil
	}

	reading.UpdatedAt = time.Now()
	r.readings[reading.ID] = *reading
	return nil
}

func (r *ReadingRepository) ListByModuleID(ctx context.Context, moduleID int64) ([]domain.Reading, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Reading, 0)
	for _, reading := range r.readings {
		if reading.ModuleID == moduleID {
			result = append(result, reading)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})

	return result, nil
}

func (r *ReadingRepository) DeleteByID(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.readings, id)
	return nil
}
