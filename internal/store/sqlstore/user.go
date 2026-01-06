package sqlstore

import (
	"bytecourses/internal/domain"
	"context"
	"database/sql"
)

func (s *Store) CreateUser(ctx context.Context, u *domain.User) error {
	return s.DB.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, role)
         this`,
	)
}
