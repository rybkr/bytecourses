package sqlstore

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"time"
)

func (s *Store) CreateModule(ctx context.Context, m *domain.Module) error {
	now := time.Now().UTC()

	var maxPos int
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(position), 0) FROM modules WHERE course_id = $1
	`, m.CourseID).Scan(&maxPos)
	if err != nil {
		return err
	}

	m.Position = maxPos + 1

	if err := s.db.QueryRowContext(ctx, `
		INSERT INTO modules (
			course_id,
			title,
			position,
			created_at, updated_at
		) VALUES (
			$1,
			$2,
			$3,
			$4, $5
		)
		RETURNING id
	`,
		m.CourseID,
		m.Title,
		m.Position,
		now, now,
	).Scan(&m.ID); err != nil {
		return err
	}

	m.CreatedAt = now
	m.UpdatedAt = now
	return nil
}

func (s *Store) GetModuleByID(ctx context.Context, id int64) (*domain.Module, bool) {
	var m domain.Module

	if err := s.db.QueryRowContext(ctx, `
		SELECT id, course_id,
		       title,
		       position,
		       created_at, updated_at
		  FROM modules
		 WHERE id = $1
	`, id).Scan(
		&m.ID, &m.CourseID,
		&m.Title,
		&m.Position,
		&m.CreatedAt, &m.UpdatedAt,
	); err != nil {
		return nil, false
	}

	return &m, true
}

func (s *Store) UpdateModule(ctx context.Context, m *domain.Module) error {
	now := time.Now().UTC()

	res, err := s.db.ExecContext(ctx, `
		UPDATE modules
		   SET title = $2,
		       position = $3,
		       updated_at = $4
		 WHERE id = $1
	`,
		m.ID,
		m.Title,
		m.Position,
		now,
	)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return store.ErrNotFound
	}

	m.UpdatedAt = now
	return nil
}

func (s *Store) DeleteModuleByID(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM modules WHERE id = $1`, id)
	if err != nil {
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

func (s *Store) ListModulesByCourseID(ctx context.Context, courseID int64) ([]domain.Module, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, course_id,
		       title,
		       position,
		       created_at, updated_at
		  FROM modules
		 WHERE course_id = $1
		 ORDER BY position ASC, id ASC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Module, 0, 16)
	for rows.Next() {
		var m domain.Module

		if err := rows.Scan(
			&m.ID, &m.CourseID,
			&m.Title,
			&m.Position,
			&m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, err
		}

		out = append(out, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) ReorderModules(ctx context.Context, courseID int64, moduleIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	for i, id := range moduleIDs {
		res, err := tx.ExecContext(ctx, `
			UPDATE modules
			   SET position = $2,
			       updated_at = $3
			 WHERE id = $1 AND course_id = $4
		`, id, i+1, now, courseID)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			return store.ErrNotFound
		}
	}

	return tx.Commit()
}
