package sqlstore

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"github.com/jackc/pgconn"
	"time"
)

func (s *Store) CreateUser(ctx context.Context, u *domain.User) error {
	createdAt := time.Now().UTC()
    role := u.Role
    if role == "" {
        role = domain.UserRoleStudent
    }

	if err := s.db.QueryRowContext(ctx, `
        INSERT INTO users (name, email, password_hash, role, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `,
		u.Name,
		u.Email,
		u.PasswordHash,
		string(role),
		createdAt,
	).Scan(&u.ID); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return store.ErrConflict
		}
		return err
	}

	u.CreatedAt = createdAt
	return nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*domain.User, bool) {
	var u domain.User
	var role string

	if err := s.db.QueryRowContext(ctx, `
        SELECT id, name, email, password_hash, role, created_at
        FROM users
        WHERE id = $1
    `, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&role,
		&u.CreatedAt,
	); err != nil {
		return nil, false
	}

	u.Role = domain.UserRole(role)
	return &u, true
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*domain.User, bool) {
	var u domain.User
	var role string

	if err := s.db.QueryRowContext(ctx, `
		SELECT id, name, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&role,
		&u.CreatedAt,
	); err != nil {
		return nil, false
	}

	u.Role = domain.UserRole(role)
	return &u, true
}

func (s *Store) UpdateUser(ctx context.Context, u *domain.User) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE users
		   SET name = $2,
		       email = $3,
		       password_hash = $4,
		       role = $5
		 WHERE id = $1
	`,
		u.ID,
		u.Name,
		u.Email,
		u.PasswordHash,
		string(u.Role),
	)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return store.ErrConflict
		}
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}
	return nil
}
