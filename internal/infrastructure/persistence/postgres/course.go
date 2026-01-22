package postgres

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"context"
	"database/sql"
	"time"
)

var _ persistence.CourseRepository = (*CourseRepository)(nil)

type CourseRepository struct {
	db *sql.DB
}

func NewCourseRepository(db *DB) *CourseRepository {
	return &CourseRepository{db: db.DB()}
}

func (r *CourseRepository) Create(ctx context.Context, c *domain.Course) error {
	now := time.Now().UTC()

	if err := r.db.QueryRowContext(ctx, `
		INSERT INTO courses (
			title, summary, target_audience, learning_objectives,
			assumed_prerequisites, instructor_id, status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`,
		c.Title,
		c.Summary,
		c.TargetAudience,
		c.LearningObjectives,
		c.AssumedPrerequisites,
		c.InstructorID,
		string(c.Status),
		now,
		now,
	).Scan(&c.ID); err != nil {
		return err
	}

	c.CreatedAt = now
	c.UpdatedAt = now
	return nil
}

func (r *CourseRepository) GetByID(ctx context.Context, id int64) (*domain.Course, bool) {
	var c domain.Course
	var status string

	if err := r.db.QueryRowContext(ctx, `
		SELECT id, title, summary, target_audience, learning_objectives,
		       assumed_prerequisites, instructor_id, status,
		       created_at, updated_at
		FROM courses
		WHERE id = $1
	`, id).Scan(
		&c.ID,
		&c.Title,
		&c.Summary,
		&c.TargetAudience,
		&c.LearningObjectives,
		&c.AssumedPrerequisites,
		&c.InstructorID,
		&status,
		&c.CreatedAt,
		&c.UpdatedAt,
	); err != nil {
		return nil, false
	}

	c.Status = domain.CourseStatus(status)
	return &c, true
}

func (r *CourseRepository) ListAllLive(ctx context.Context) ([]domain.Course, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, summary, target_audience, learning_objectives,
		       assumed_prerequisites, instructor_id, status,
		       created_at, updated_at
		FROM courses
		WHERE status = 'live'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := make([]domain.Course, 0)
	for rows.Next() {
		var c domain.Course
		var status string

		if err := rows.Scan(
			&c.ID,
			&c.Title,
			&c.Summary,
			&c.TargetAudience,
			&c.LearningObjectives,
			&c.AssumedPrerequisites,
			&c.InstructorID,
			&status,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, err
		}

		c.Status = domain.CourseStatus(status)
		courses = append(courses, c)
	}

	return courses, rows.Err()
}

func (r *CourseRepository) Update(ctx context.Context, c *domain.Course) error {
	c.UpdatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, `
		UPDATE courses
		SET title = $2,
		    summary = $3,
		    target_audience = $4,
		    learning_objectives = $5,
		    assumed_prerequisites = $6,
		    status = $7,
		    updated_at = $8
		WHERE id = $1
	`,
		c.ID,
		c.Title,
		c.Summary,
		c.TargetAudience,
		c.LearningObjectives,
		c.AssumedPrerequisites,
		string(c.Status),
		c.UpdatedAt,
	)
	return err
}
