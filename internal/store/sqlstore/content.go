package sqlstore

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"time"
)

func (s *Store) CreateContentItem(ctx context.Context, item *domain.ContentItem) error {
	now := time.Now().UTC()

	var maxPos int
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(position), 0) FROM content_items WHERE module_id = $1
	`, item.ModuleID).Scan(&maxPos)
	if err != nil {
		return err
	}

	item.Position = maxPos + 1

	if err := s.db.QueryRowContext(ctx, `
		INSERT INTO content_items (
			module_id,
			title,
			type,
			status,
			position,
			created_at, updated_at
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6, $7
		)
		RETURNING id
	`,
		item.ModuleID,
		item.Title,
		item.Type,
		item.Status,
		item.Position,
		now, now,
	).Scan(&item.ID); err != nil {
		return err
	}

	item.CreatedAt = now
	item.UpdatedAt = now
	return nil
}

func (s *Store) GetContentItemByID(ctx context.Context, id int64) (*domain.ContentItem, bool) {
	var item domain.ContentItem

	if err := s.db.QueryRowContext(ctx, `
		SELECT id, module_id,
		       title, type, status,
		       position,
		       created_at, updated_at
		  FROM content_items
		 WHERE id = $1
	`, id).Scan(
		&item.ID, &item.ModuleID,
		&item.Title, &item.Type, &item.Status,
		&item.Position,
		&item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, false
	}

	return &item, true
}

func (s *Store) UpdateContentItem(ctx context.Context, item *domain.ContentItem) error {
	now := time.Now().UTC()

	res, err := s.db.ExecContext(ctx, `
		UPDATE content_items
		   SET title = $2,
		       status = $3,
		       position = $4,
		       updated_at = $5
		 WHERE id = $1
	`,
		item.ID,
		item.Title,
		item.Status,
		item.Position,
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

	item.UpdatedAt = now
	return nil
}

func (s *Store) DeleteContentItemByID(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM content_items WHERE id = $1`, id)
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

func (s *Store) ListContentItemsByModuleID(ctx context.Context, moduleID int64) ([]domain.ContentItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, module_id,
		       title, type, status,
		       position,
		       created_at, updated_at
		  FROM content_items
		 WHERE module_id = $1
		 ORDER BY position ASC, id ASC
	`, moduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.ContentItem, 0, 16)
	for rows.Next() {
		var item domain.ContentItem

		if err := rows.Scan(
			&item.ID, &item.ModuleID,
			&item.Title, &item.Type, &item.Status,
			&item.Position,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}

		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) ReorderContentItems(ctx context.Context, moduleID int64, itemIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC()

	for i, id := range itemIDs {
		res, err := tx.ExecContext(ctx, `
			UPDATE content_items
			   SET position = $2,
			       updated_at = $3
			 WHERE id = $1 AND module_id = $4
		`, id, i+1, now, moduleID)
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

func (s *Store) GetLecture(ctx context.Context, contentItemID int64) (*domain.Lecture, bool) {
	var lecture domain.Lecture

	if err := s.db.QueryRowContext(ctx, `
		SELECT content_item_id, content
		  FROM lectures
		 WHERE content_item_id = $1
	`, contentItemID).Scan(
		&lecture.ContentItemID, &lecture.Content,
	); err != nil {
		return nil, false
	}

	return &lecture, true
}

func (s *Store) UpsertLecture(ctx context.Context, lecture *domain.Lecture) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO lectures (content_item_id, content)
		VALUES ($1, $2)
		ON CONFLICT (content_item_id)
		DO UPDATE SET content = EXCLUDED.content
	`, lecture.ContentItemID, lecture.Content)
	return err
}

func (s *Store) GetContentItemWithLecture(ctx context.Context, id int64) (*domain.ContentItem, *domain.Lecture, bool) {
	var item domain.ContentItem
	var lecture domain.Lecture
	var lectureContent *string

	if err := s.db.QueryRowContext(ctx, `
		SELECT ci.id, ci.module_id,
		       ci.title, ci.type, ci.status,
		       ci.position,
		       ci.created_at, ci.updated_at,
		       l.content
		  FROM content_items ci
		  LEFT JOIN lectures l ON ci.id = l.content_item_id
		 WHERE ci.id = $1
	`, id).Scan(
		&item.ID, &item.ModuleID,
		&item.Title, &item.Type, &item.Status,
		&item.Position,
		&item.CreatedAt, &item.UpdatedAt,
		&lectureContent,
	); err != nil {
		return nil, nil, false
	}

	var lecPtr *domain.Lecture
	if lectureContent != nil {
		lecture.ContentItemID = item.ID
		lecture.Content = *lectureContent
		lecPtr = &lecture
	}

	return &item, lecPtr, true
}

func (s *Store) ListContentItemsWithLecturesByModuleID(ctx context.Context, moduleID int64) ([]domain.ContentItem, map[int64]*domain.Lecture, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ci.id, ci.module_id,
		       ci.title, ci.type, ci.status,
		       ci.position,
		       ci.created_at, ci.updated_at,
		       l.content
		  FROM content_items ci
		  LEFT JOIN lectures l ON ci.id = l.content_item_id
		 WHERE ci.module_id = $1
		 ORDER BY ci.position ASC, ci.id ASC
	`, moduleID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	items := make([]domain.ContentItem, 0, 16)
	lectures := make(map[int64]*domain.Lecture)

	for rows.Next() {
		var item domain.ContentItem
		var lectureContent *string

		if err := rows.Scan(
			&item.ID, &item.ModuleID,
			&item.Title, &item.Type, &item.Status,
			&item.Position,
			&item.CreatedAt, &item.UpdatedAt,
			&lectureContent,
		); err != nil {
			return nil, nil, err
		}

		items = append(items, item)

		if lectureContent != nil {
			lectures[item.ID] = &domain.Lecture{
				ContentItemID: item.ID,
				Content:       *lectureContent,
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return items, lectures, nil
}
