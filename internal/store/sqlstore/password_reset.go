package sqlstore

import (
	"context"
	"errors"
	"time"
)

func (s *Store) CreateResetToken(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	if userID <= 0 {
		return errors.New("invalid user id")
	}
	if len(tokenHash) == 0 {
		return errors.New("missing token hash")
	}
	if expiresAt.IsZero() {
		return errors.New("missing expiresAt")
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO password_reset_tokens (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, tokenHash, userID, expiresAt)
	return err
}

func (s *Store) ConsumeResetToken(ctx context.Context, tokenHash []byte, now time.Time) (int64, bool) {
	if len(tokenHash) == 0 {
		return 0, false
	}
	if now.IsZero() {
		now = time.Now()
	}

	var userID int64
	if err := s.db.QueryRowContext(ctx, `
		UPDATE password_reset_tokens
		SET used_at = $2
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > $2
		RETURNING user_id
	`, tokenHash, now).Scan(&userID); err != nil {
		return 0, false
	}
	return userID, true
}
