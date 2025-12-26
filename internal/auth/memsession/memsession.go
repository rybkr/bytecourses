package memsession

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

type Session struct {
	userID  int64
	expires time.Time
}

type Store struct {
	mu      sync.Mutex
	ttl     time.Duration
	byToken map[string]Session
}

func New(ttl time.Duration) *Store {
	return &Store{
		ttl:     ttl,
		byToken: make(map[string]Session),
	}
}

func (s *Store) InsertSession(userID int64) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}

	token := base64.RawURLEncoding.EncodeToString(b)
	exp := time.Now().Add(s.ttl)

	s.mu.Lock()
	s.byToken[token] = Session{
		userID:  userID,
		expires: exp,
	}
	s.mu.Unlock()

	return token, exp, nil
}

func (s *Store) GetUserIDByToken(token string) (int64, bool) {
    now := time.Now()
    s.mu.Lock()
    defer s.mu.Unlock()

    session, ok := s.byToken[token]
    if !ok || now.After(session.expires) {
        delete(s.byToken, token)
        return 0, false
    }

    return session.userID, true
}

func (s *Store) DeleteSession(token string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.byToken, token)
}
