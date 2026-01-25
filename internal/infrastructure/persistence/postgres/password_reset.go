package postgres

import (
	"context"
	"database/sql"
	"time"
)

type PasswordResetRepository struct {
	db *sql.DB
}

func NewPasswordResetRepository(db *DB) *PasswordResetRepository {
	return &PasswordResetRepository{db: db.DB()}
}

func (r *PasswordResetRepository) CreateResetToken(ctx context.Context, userID int64, tokenHash []byte, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *PasswordResetRepository) ConsumeResetToken(ctx context.Context, tokenHash []byte, now time.Time) (int64, bool) {
	var userID int64

	err := r.db.QueryRowContext(ctx, `
		DELETE FROM password_reset_tokens
		WHERE token_hash = $1
		  AND expires_at > $2
		  AND used_at IS NULL
		RETURNING user_id
	`, tokenHash, now).Scan(&userID)

	if err != nil {
		return 0, false
	}

	return userID, true
}
