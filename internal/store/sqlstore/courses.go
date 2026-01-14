package sqlstore

import (
	"bytecourses/internal/domain"
	"context"
	"database/sql"
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
            proposal_id,
            status,
            created_at
        ) VALUES (
            $1,
            $2, $3,
            $4,
            $5,
            $6
        )
        RETURNING id
    `,
		c.InstructorID,
		c.Title, c.Summary,
		nullInt64Ptr(c.ProposalID),
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
	var proposalID sql.NullInt64

	if err := s.db.QueryRowContext(ctx, `
        SELECT id, instructor_id,
               title, summary,
               proposal_id,
               status,
               created_at
          FROM courses
         WHERE id = $1
    `, id).Scan(
		&c.ID, &c.InstructorID,
		&c.Title, &c.Summary,
		&proposalID,
		&status,
		&c.CreatedAt,
	); err != nil {
		return nil, false
	}

	c.Status = domain.CourseStatus(status)
	c.ProposalID = ptrFromNullInt64(proposalID)
	return &c, true
}

func (s *Store) GetCourseByProposalID(ctx context.Context, proposalID int64) (*domain.Course, bool) {
	var c domain.Course
	var status string
	var pid sql.NullInt64

	if err := s.db.QueryRowContext(ctx, `
        SELECT id, instructor_id,
               title, summary,
               proposal_id,
               status,
               created_at
          FROM courses
         WHERE proposal_id = $1
    `, proposalID).Scan(
		&c.ID, &c.InstructorID,
		&c.Title, &c.Summary,
		&pid,
		&status,
		&c.CreatedAt,
	); err != nil {
		return nil, false
	}

	c.Status = domain.CourseStatus(status)
	c.ProposalID = ptrFromNullInt64(pid)
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
