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
	mu         sync.RWMutex
	usersByID  map[int64]domain.User
	idsByEmail map[string]int64
	nextID     int64
}

func NewUserStore() *UserStore {
	return &UserStore{
		usersByID:  make(map[int64]domain.User),
		idsByEmail: make(map[string]int64),
		nextID:     1,
	}
}

func (s *UserStore) InsertUser(ctx context.Context, u *domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u.Email = normalizeEmailAddress(u.Email)
	if u.Email == "" {
		return errors.New("email required")
	}
	if _, exists := s.idsByEmail[u.Email]; exists {
		return errors.New("email already exists")
	}

	u.ID = s.nextID
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}

	s.usersByID[u.ID] = *u
	s.idsByEmail[u.Email] = u.ID
	s.nextID++

	return nil
}

func (s *UserStore) GetUserByID(ctx context.Context, id int64) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.usersByID[id]
	return u, ok
}

func (s *UserStore) GetUserByEmail(ctx context.Context, email string) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uid, ok := s.idsByEmail[normalizeEmailAddress(email)]
	if !ok {
		return domain.User{}, false
	}

	return s.GetUserByID(ctx, uid)
}

func (s *UserStore) UpdateUser(ctx context.Context, u *domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u.Email = normalizeEmailAddress(u.Email)
	if _, exists := s.usersByID[u.ID]; !exists {
		return errors.New("user does not exist")
	}

	s.usersByID[u.ID] = *u
	s.idsByEmail[u.Email] = u.ID

	return nil
}

func normalizeEmailAddress(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
