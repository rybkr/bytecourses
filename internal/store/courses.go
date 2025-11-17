package store

import (
	"context"
	"github.com/rybkr/bytecourses/internal/models"
)

func (s *Store) CreateCourse(ctx context.Context, course *models.Course) error {
	query := `
        INSERT INTO courses (instructor_id, title, description)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`

	return s.db.QueryRow(ctx, query,
		course.InstructorID,
		course.Title,
		course.Description,
	).Scan(&course.ID, &course.CreatedAt, &course.UpdatedAt)
}

func (s *Store) GetCourses(ctx context.Context) ([]*models.Course, error) {
	query := `
        SELECT id, instructor_id, title, description, status, created_at, updated_at
        FROM courses
        ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []*models.Course
	for rows.Next() {
		var c models.Course
		err := rows.Scan(
			&c.ID, &c.InstructorID, &c.Title, &c.Description,
			&c.Status, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		courses = append(courses, &c)
	}
	return courses, rows.Err()
}

func (s *Store) UpdateCourseStatus(ctx context.Context, courseID int, status models.CourseStatus) error {
	query := `UPDATE courses SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := s.db.Exec(ctx, query, status, courseID)
	return err
}
