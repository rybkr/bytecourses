package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
)

var (
	_ persistence.UserRepository = (*UserRepository)(nil)
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{
		db: db.DB(),
	}
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	createdAt := time.Now().UTC()
	role := u.Role
	if role == "" {
		role = domain.UserRoleStudent
	}

	if err := r.db.QueryRowContext(ctx, `
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
			return err
		}
		return err
	}

	u.CreatedAt = createdAt
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, bool) {
	var u domain.User
	var role string

	if err := r.db.QueryRowContext(ctx, `
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

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, bool) {
	var u domain.User
	var role string

	if err := r.db.QueryRowContext(ctx, `
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

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	result, err := r.db.ExecContext(ctx, `
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
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}

func (r *UserRepository) DeleteByID(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM users
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}
