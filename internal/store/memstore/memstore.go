package memstore

import (
	"bytecourses/internal/domain"
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

type Store struct {
	mu sync.Mutex

	usersByID    map[int64]*domain.User
	usersByEmail map[string]*domain.User
	nextUserID   int64
}

func New() *Store {
	return &Store{
		usersByID:      make(map[int64]domain.User),
		userIDsByEmail: make(map[string]int64),
		proposalsByID:  make(map[int64]domain.CourseProposal),
		nextUserID:     1,
	}
}

func (s *Store) CreateUser(ctx context.Context, u *domain.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := strings.ToLower(strings.TrimSpace(u.Email))
	if key == "" {
		return errors.New("email required")
	}
	if _, exists := s.userIDsByEmail[key]; exists {
		return errors.New("email already exists")
	}

	u.ID = s.nextUserID
	u.Email = key
	if u.CreatedAt.IsZero() {
		u.CreatedAt = time.Now()
	}

	s.usersByID[u.ID] = u
	s.userIDsByEmail[u.Email] = u.ID
	s.nextUserID++

	return nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*domain.User, bool, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    u, ok := s.usersByID[id]
    return u, ok, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*domain.User, bool, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    key := strings.ToLower(strings.TrimSpace(email))
    u, ok := s.usersByEmail[key]
    return u, ok, nil
}
