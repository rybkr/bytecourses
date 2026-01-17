package memsession

import (
	"bytecourses/internal/auth"
	"sync"
	"time"
)

type Session struct {
	userID    int64
	expiresAt time.Time
}

type Store struct {
	mu      sync.RWMutex
	ttl     time.Duration
	byToken map[string]Session
}

func New(ttl time.Duration) *Store {
	return &Store{
		ttl:     ttl,
		byToken: make(map[string]Session),
	}
}

func (s *Store) CreateSession(userID int64) (string, error) {
	for {
		token, err := auth.GenerateToken(32)
		if err != nil {
			return "", err
		}
		exp := time.Now().Add(s.ttl)

		s.mu.Lock()
		if _, exists := s.byToken[token]; !exists {
			s.byToken[token] = Session{
				userID:    userID,
				expiresAt: exp,
			}
			s.mu.Unlock()
			return token, nil
		}
		s.mu.Unlock()
	}
}

func (s *Store) GetUserIDByToken(token string) (int64, bool) {
	now := time.Now()

	s.mu.RLock()
	session, ok := s.byToken[token]
	s.mu.RUnlock()

	if !ok {
		return 0, false
	}
	if now.Before(session.expiresAt) {
		return session.userID, true
	}

	s.mu.Lock()
	session, ok = s.byToken[token]
	if ok && now.After(session.expiresAt) {
		delete(s.byToken, token)
		s.mu.Unlock()
		return 0, false
	}
	s.mu.Unlock()

	if !ok {
		return 0, false
	}
	return session.userID, true
}

func (s *Store) DeleteSessionByToken(token string) {
	s.mu.Lock()
	delete(s.byToken, token)
	s.mu.Unlock()
}
