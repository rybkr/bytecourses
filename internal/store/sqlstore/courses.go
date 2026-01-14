package sqlstore

import (
	"bytecourses/internal/domain"
	"context"
	"time"
)

func (s *Store) CreateCourse(ctx context.Context, c *domain.Course) error {
	now := time.Now().UTC()
	status := c.Status
	if status == "" {
		status = domain.CourseStatusDraft
	}

	if err := s.db.QueryRowContext(ctx, `
        INSERT INTO courses (
            instructor_id,
            title, summary,
            status,
            created_at
        ) VALUES (
            $1,
            $2, $3,
            $4,
            $5
        )
        RETURNING id
    `,
		c.InstructorID,
		c.Title, c.Summary,
		string(status),
		now,
	).Scan(&c.ID); err != nil {
		return err
	}

	c.CreatedAt = now
	return nil
}

func (s *Store) GetCourseByID(ctx context.Context, id int64) (*domain.Course, bool) {
	var c domain.Course
	var status string

	if err := s.db.QueryRowContext(ctx, `
        SELECT id, instructor_id,
               title, summary,
               status,
               created_at
          FROM courses
         WHERE id = $1
    `, id).Scan(
		&c.ID, &c.InstructorID,
		&c.Title, &c.Summary,
		&status,
		&c.CreatedAt,
	); err != nil {
		return nil, false
	}

	c.Status = domain.CourseStatus(status)
	return &c, true
}

func (s *Store) ListAllLiveCourses(ctx context.Context) ([]domain.Course, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT id, instructor_id,
               title, summary,
               status,
               created_at
          FROM courses
         WHERE status IN ('live')
          ORDER BY created_at DESC, id DESC
    `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Course, 0, 16)
	for rows.Next() {
		var c domain.Course
		var status string

		if err := rows.Scan(
			&c.ID, &c.InstructorID,
			&c.Title, &c.Summary,
			&status,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}

		c.Status = domain.CourseStatus(status)
		out = append(out, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
