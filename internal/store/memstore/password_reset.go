package memstore

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrResetTokenExists = errors.New("reset token already exists")
)

type resetRow struct {
	userID    int64
	createdAt time.Time
	expiresAt time.Time
	usedAt    *time.Time
}

type PasswordResetStore struct {
	mu              sync.Mutex
	resetRowsByHash map[string]resetRow
}

func NewPasswordResetStore() *PasswordResetStore {
	return &PasswordResetStore{
		resetRowsByHash: make(map[string]resetRow),
	}
}

func (s *PasswordResetStore) CreateResetToken(_ context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	if userID <= 0 || len(tokenHash) == 0 {
		return errors.New("invalid reset token")
	}
	key := string(tokenHash)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.resetRowsByHash[key]; exists {
		return ErrResetTokenExists
	}

	s.resetRowsByHash[key] = resetRow{
		userID:    userID,
		expiresAt: expiresAt,
		createdAt: time.Now(),
	}
	return nil
}

func (s *PasswordResetStore) ConsumeResetToken(_ context.Context, tokenHash []byte, now time.Time) (int64, bool) {
	if len(tokenHash) == 0 {
		return 0, false
	}
	key := string(tokenHash)

	s.mu.Lock()
	defer s.mu.Unlock()

	row, ok := s.resetRowsByHash[key]
	if !ok {
		return 0, false
	}
	if row.usedAt != nil || !now.Before(row.expiresAt) {
		delete(s.resetRowsByHash, key)
		return 0, false
	}

	t := now
	row.usedAt = &t
	s.resetRowsByHash[key] = row

	return row.userID, true
}
