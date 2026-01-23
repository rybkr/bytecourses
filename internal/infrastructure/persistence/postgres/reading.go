package postgres

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"context"
	"database/sql"
	"time"
)

var _ persistence.ReadingRepository = (*ReadingRepository)(nil)

type ReadingRepository struct {
	db *sql.DB
}

func NewReadingRepository(db *DB) *ReadingRepository {
	return &ReadingRepository{db: db.DB()}
}

func (r *ReadingRepository) Create(ctx context.Context, reading *domain.Reading) error {
	now := time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var contentItemID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO content_items (
			module_id, content_type, title, order_index, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`,
		reading.ModuleID,
		string(reading.Type()),
		reading.Title,
		reading.Order,
		string(reading.Status),
		now,
		now,
	).Scan(&contentItemID); err != nil {
		return err
	}

	var content sql.NullString
	if reading.Content != nil {
		content.String = *reading.Content
		content.Valid = true
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO readings (
			content_item_id, format, content
		)
		VALUES ($1, $2, $3)
	`,
		contentItemID,
		string(reading.Format),
		content,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	reading.ID = contentItemID
	reading.CreatedAt = now
	reading.UpdatedAt = now
	return nil
}

func (r *ReadingRepository) GetByID(ctx context.Context, id int64) (*domain.Reading, bool) {
	var reading domain.Reading
	var status string
	var contentType string
	var format string
	var content sql.NullString

	if err := r.db.QueryRowContext(ctx, `
		SELECT ci.id, ci.module_id, ci.content_type, ci.title, ci.order_index,
		       ci.status, ci.created_at, ci.updated_at,
		       r.format, r.content
		FROM content_items ci
		INNER JOIN readings r ON ci.id = r.content_item_id
		WHERE ci.id = $1
	`, id).Scan(
		&reading.ID,
		&reading.ModuleID,
		&contentType,
		&reading.Title,
		&reading.Order,
		&status,
		&reading.CreatedAt,
		&reading.UpdatedAt,
		&format,
		&content,
	); err != nil {
		return nil, false
	}

	if contentType != string(domain.ContentTypeReading) {
		return nil, false
	}

	reading.Status = domain.ContentStatus(status)
	reading.Format = domain.ReadingFormat(format)
	if content.Valid {
		reading.Content = &content.String
	}

	return &reading, true
}

func (r *ReadingRepository) Update(ctx context.Context, reading *domain.Reading) error {
	reading.UpdatedAt = time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE content_items
		SET title = $2,
		    order_index = $3,
		    status = $4,
		    updated_at = $5
		WHERE id = $1
	`,
		reading.ID,
		reading.Title,
		reading.Order,
		string(reading.Status),
		reading.UpdatedAt,
	)
	if err != nil {
		return err
	}

	var content sql.NullString
	if reading.Content != nil {
		content.String = *reading.Content
		content.Valid = true
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE readings
		SET format = $2,
		    content = $3
		WHERE content_item_id = $1
	`,
		reading.ID,
		string(reading.Format),
		content,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ReadingRepository) ListByModuleID(ctx context.Context, moduleID int64) ([]domain.Reading, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ci.id, ci.module_id, ci.content_type, ci.title, ci.order_index,
		       ci.status, ci.created_at, ci.updated_at,
		       r.format, r.content
		FROM content_items ci
		INNER JOIN readings r ON ci.id = r.content_item_id
		WHERE ci.module_id = $1
		ORDER BY ci.order_index ASC
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReadings(rows)
}

func (r *ReadingRepository) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM readings WHERE content_item_id = $1`, id)
	return err
}

func scanReadings(rows *sql.Rows) ([]domain.Reading, error) {
	readings := make([]domain.Reading, 0)

	for rows.Next() {
		var reading domain.Reading
		var status string
		var contentType string
		var format string
		var content sql.NullString

		if err := rows.Scan(
			&reading.ID,
			&reading.ModuleID,
			&contentType,
			&reading.Title,
			&reading.Order,
			&status,
			&reading.CreatedAt,
			&reading.UpdatedAt,
			&format,
			&content,
		); err != nil {
			return nil, err
		}

		if contentType != string(domain.ContentTypeReading) {
			continue
		}

		reading.Status = domain.ContentStatus(status)
		reading.Format = domain.ReadingFormat(format)
		if content.Valid {
			reading.Content = &content.String
		}

		readings = append(readings, reading)
	}

	return readings, rows.Err()
}
