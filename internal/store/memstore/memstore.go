package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

type UserStore struct {
	mu sync.RWMutex

	byID    map[int64]*domain.User
	byEmail map[string]*domain.User
	nextID  int64
}

func NewUserStore() *UserStore {
	return &UserStore{
		byID:    make(map[int64]*domain.User),
		byEmail: make(map[string]*domain.User),
		nextID:  1,
	}
}

// CreateUser persists a new user and assigns a unique ID.
// The store takes ownership of u. The caller must not mutatue u after calling.
func (s *UserStore) CreateUser(ctx context.Context, u *domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	email := strings.ToLower(strings.TrimSpace(u.Email))
	if email == "" {
		return errors.New("email required")
	}
	if _, exists := s.byEmail[email]; exists {
		return errors.New("email already exists")
	}

	u.ID = s.nextID
	u.Email = email
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}

	s.byID[u.ID] = u
	s.byEmail[u.Email] = u
	s.nextID++

	return nil
}

// GetUserByID returns the user with the given ID.
func (s *UserStore) GetUserByID(ctx context.Context, id int64) (*domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[id]
	return u, ok
}

// GetUserByEmail returns the user with the given email.
func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (*domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
    u, ok := s.byEmail[strings.ToLower(strings.TrimSpace(email))]
	return u, ok
}
