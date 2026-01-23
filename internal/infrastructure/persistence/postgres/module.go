package postgres

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"context"
	"database/sql"
	"time"
)

var _ persistence.ModuleRepository = (*ModuleRepository)(nil)

type ModuleRepository struct {
	db *sql.DB
}

func NewModuleRepository(db *DB) *ModuleRepository {
	return &ModuleRepository{db: db.DB()}
}

func (r *ModuleRepository) Create(ctx context.Context, m *domain.Module) error {
	now := time.Now().UTC()

	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO modules (
			course_id, title, description, order_index, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`,
		m.CourseID,
		m.Title,
		m.Description,
		m.Order,
		string(m.Status),
		now,
		now,
	).Scan(&m.ID); err != nil {
		return err
	}

	m.CreatedAt = now
	m.UpdatedAt = now
	return nil
}

func (r *ModuleRepository) GetByID(ctx context.Context, id int64) (*domain.Module, bool) {
	var m domain.Module
	var status string

	if err := r.db.QueryRowContext(ctx, `
		SELECT id, course_id, title, description, order_index, status,
		       created_at, updated_at
		FROM modules
		WHERE id = $1
	`, id).Scan(
		&m.ID,
		&m.CourseID,
		&m.Title,
		&m.Description,
		&m.Order,
		&status,
		&m.CreatedAt,
		&m.UpdatedAt,
	); err != nil {
		return nil, false
	}

	m.Status = domain.ModuleStatus(status)
	return &m, true
}

func (r *ModuleRepository) Update(ctx context.Context, m *domain.Module) error {
	m.UpdatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		UPDATE modules
		SET title = $2,
		    description = $3,
		    order_index = $4,
		    status = $5,
		    updated_at = $6
		WHERE id = $1
	`,
		m.ID,
		m.Title,
		m.Description,
		m.Order,
		string(m.Status),
		m.UpdatedAt,
	)
	return err
}

func (r *ModuleRepository) ListByCourseID(ctx context.Context, courseID int64) ([]domain.Module, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, course_id, title, description, order_index, status,
		       created_at, updated_at
		FROM modules
		WHERE course_id = $1
		ORDER BY order_index ASC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanModules(rows)
}

func (r *ModuleRepository) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM modules WHERE id = $1`, id)
	return err
}

func scanModules(rows *sql.Rows) ([]domain.Module, error) {
	modules := make([]domain.Module, 0)

	for rows.Next() {
		var m domain.Module
		var status string

		if err := rows.Scan(
			&m.ID,
			&m.CourseID,
			&m.Title,
			&m.Description,
			&m.Order,
			&status,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, err
		}

		m.Status = domain.ModuleStatus(status)
		modules = append(modules, m)
	}

	return modules, rows.Err()
}
