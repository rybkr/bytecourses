package memory

import (
	"context"
	"sync"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/infrastructure/persistence"
)

var (
	_ persistence.UserRepository = (*UserRepository)(nil)
)

type UserRepository struct {
	mu      sync.RWMutex
	users   map[int64]domain.User
	byEmail map[string]int64
	nextID  int64
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		users:   make(map[int64]domain.User),
		byEmail: make(map[string]int64),
		nextID:  1,
	}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

    if _, exists := r.byEmail[u.Email]; exists {
        return errors.ErrConflict
    }

	u.ID = r.nextID
	r.nextID++
	u.CreatedAt = time.Now()

	r.users[u.ID] = *u
	r.byEmail[u.Email] = u.ID

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, false
	}

	return &u, true
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byEmail[email]
	if !ok {
		return nil, false
	}

	return r.GetByID(ctx, id)
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.users[u.ID]
	if !ok {
		return errors.ErrNotFound
	}

	if existing.Email != u.Email {
        if _, exists := r.byEmail[u.Email]; exists {
            return errors.ErrConflict
        }
		delete(r.byEmail, existing.Email)
		r.byEmail[u.Email] = u.ID
	}

	r.users[u.ID] = *u

	return nil
}
