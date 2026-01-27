package memory

import (
	"context"
	"sort"
	"sync"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

var _ persistence.FileRepository = (*FileRepository)(nil)

type FileRepository struct {
	mu     sync.RWMutex
	files  map[int64]domain.File
	nextID int64
}

func NewFileRepository() *FileRepository {
	return &FileRepository{
		files:  make(map[int64]domain.File),
		nextID: 1,
	}
}

func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	file.ID = r.nextID
	r.nextID++
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()

	r.files[file.ID] = *file
	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id int64) (*domain.File, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	file, ok := r.files[id]
	if !ok {
		return nil, false
	}

	return &file, true
}

func (r *FileRepository) Update(ctx context.Context, file *domain.File) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.files[file.ID]; !ok {
		return nil
	}

	file.UpdatedAt = time.Now()
	r.files[file.ID] = *file
	return nil
}

func (r *FileRepository) ListByModuleID(ctx context.Context, moduleID int64) ([]domain.File, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.File, 0)
	for _, file := range r.files {
		if file.ModuleID == moduleID {
			result = append(result, file)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})

	return result, nil
}

func (r *FileRepository) DeleteByID(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.files, id)
	return nil
}
