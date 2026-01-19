package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type SessionStore interface {
	Create(userID int64) (sessionID string, err error)
	Get(sessionID string) (userID int64, ok bool)
	Delete(sessionID string) error
}

var (
	_ SessionStore = (*InMemorySessionStore)(nil)
)

type session struct {
	userID    int64
	expiresAt time.Time
}

type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]session
	ttl      time.Duration
}

func NewInMemorySessionStore(ttl time.Duration) *InMemorySessionStore {
	store := &InMemorySessionStore{
		sessions: make(map[string]session),
		ttl:      ttl,
	}
	go store.cleanup()
	return store
}

func (s *InMemorySessionStore) Create(userID int64) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	sessionID := hex.EncodeToString(tokenBytes)

	s.mu.Lock()
	s.sessions[sessionID] = session{
		userID:    userID,
		expiresAt: time.Now().Add(s.ttl),
	}
	s.mu.Unlock()

	return sessionID, nil
}

func (s *InMemorySessionStore) Get(sessionID string) (int64, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[sessionID]
	s.mu.RUnlock()

	if !ok || time.Now().After(sess.expiresAt) {
		return 0, false
	}
	return sess.userID, true
}

func (s *InMemorySessionStore) Delete(sessionID string) error {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
	return nil
}

func (s *InMemorySessionStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, sess := range s.sessions {
			if now.After(sess.expiresAt) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}
