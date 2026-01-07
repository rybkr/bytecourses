package memsession

import (
	"crypto/rand"
	"encoding/base64"
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
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}

		token := base64.RawURLEncoding.EncodeToString(b)
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
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.byToken[token]
	if !ok || time.Now().After(session.expiresAt) {
		delete(s.byToken, token)
		return 0, false
	}

	return session.userID, true
}

func (s *Store) DeleteSessionByToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byToken, token)
}
