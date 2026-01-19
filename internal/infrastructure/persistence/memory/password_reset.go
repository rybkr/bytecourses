package memory

import (
	"bytes"
	"context"
	"sync"
	"time"
)

type resetToken struct {
	userID    int64
	tokenHash []byte
	expiresAt time.Time
	consumed  bool
}

type PasswordResetRepository struct {
	mu     sync.RWMutex
	tokens []resetToken
}

func NewPasswordResetRepository() *PasswordResetRepository {
	return &PasswordResetRepository{
		tokens: make([]resetToken, 0),
	}
}

func (r *PasswordResetRepository) CreateResetToken(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tokens = append(r.tokens, resetToken{
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		consumed:  false,
	})

	return nil
}

func (r *PasswordResetRepository) ConsumeResetToken(ctx context.Context, tokenHash []byte, now time.Time) (int64, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.tokens {
		t := &r.tokens[i]
		if !t.consumed && bytes.Equal(t.tokenHash, tokenHash) && now.Before(t.expiresAt) {
			t.consumed = true
			return t.userID, true
		}
	}

	return 0, false
}
