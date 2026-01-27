package postgres

import (
	"context"
	"database/sql"
	"time"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
)

var _ persistence.FileRepository = (*FileRepository)(nil)

type FileRepository struct {
	db *sql.DB
}

func NewFileRepository(db *DB) *FileRepository {
	return &FileRepository{db: db.DB()}
}

func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	now := time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var contentItemID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO content (
			module_id, content_type, title, order_index, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`,
		file.ModuleID,
		string(file.Type()),
		file.Title,
		file.Order,
		string(file.Status),
		now,
		now,
	).Scan(&contentItemID); err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO files (
			content_item_id, file_name, file_size, mime_type, storage_path
		)
		VALUES ($1, $2, $3, $4, $5)
	`,
		contentItemID,
		file.FileName,
		file.FileSize,
		file.MimeType,
		file.StoragePath,
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	file.ID = contentItemID
	file.CreatedAt = now
	file.UpdatedAt = now
	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id int64) (*domain.File, bool) {
	var file domain.File
	var status string
	var contentType string

	if err := r.db.QueryRowContext(ctx, `
		SELECT ci.id, ci.module_id, ci.content_type, ci.title, ci.order_index,
		       ci.status, ci.created_at, ci.updated_at,
		       f.file_name, f.file_size, f.mime_type, f.storage_path
		FROM content ci
		INNER JOIN files f ON ci.id = f.content_item_id
		WHERE ci.id = $1
	`, id).Scan(
		&file.ID,
		&file.ModuleID,
		&contentType,
		&file.Title,
		&file.Order,
		&status,
		&file.CreatedAt,
		&file.UpdatedAt,
		&file.FileName,
		&file.FileSize,
		&file.MimeType,
		&file.StoragePath,
	); err != nil {
		return nil, false
	}

	if contentType != string(domain.ContentTypeFile) {
		return nil, false
	}

	file.Status = domain.ContentStatus(status)
	return &file, true
}

func (r *FileRepository) Update(ctx context.Context, file *domain.File) error {
	file.UpdatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		UPDATE content
		SET title = $2,
		    order_index = $3,
		    status = $4,
		    updated_at = $5
		WHERE id = $1
	`,
		file.ID,
		file.Title,
		file.Order,
		string(file.Status),
		file.UpdatedAt,
	)
	return err
}

func (r *FileRepository) ListByModuleID(ctx context.Context, moduleID int64) ([]domain.File, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ci.id, ci.module_id, ci.content_type, ci.title, ci.order_index,
		       ci.status, ci.created_at, ci.updated_at,
		       f.file_name, f.file_size, f.mime_type, f.storage_path
		FROM content ci
		INNER JOIN files f ON ci.id = f.content_item_id
		WHERE ci.module_id = $1
		ORDER BY ci.order_index ASC
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	files := make([]domain.File, 0)
	for rows.Next() {
		var file domain.File
		var status string
		var contentType string

		if err := rows.Scan(
			&file.ID,
			&file.ModuleID,
			&contentType,
			&file.Title,
			&file.Order,
			&status,
			&file.CreatedAt,
			&file.UpdatedAt,
			&file.FileName,
			&file.FileSize,
			&file.MimeType,
			&file.StoragePath,
		); err != nil {
			return nil, err
		}

		if contentType != string(domain.ContentTypeFile) {
			continue
		}

		file.Status = domain.ContentStatus(status)
		files = append(files, file)
	}

	return files, rows.Err()
}

func (r *FileRepository) DeleteByID(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM content WHERE id = $1`, id)
	return err
}
