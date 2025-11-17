package store

import (
	"context"
	"github.com/rybkr/bytecourses/internal/models"
	"log"
)

func (s *Store) CreateCourse(ctx context.Context, course *models.Course) error {
	query := `
        INSERT INTO courses (instructor_id, title, description)
        VALUES ($1, $2, $3)
        RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query,
		course.InstructorID,
		course.Title,
		course.Description,
	).Scan(&course.ID, &course.CreatedAt, &course.UpdatedAt)

	if err != nil {
		log.Printf("failed to create course: %v", err)
		return err
	}

	log.Printf("course created: id=%d, title=%s", course.ID, course.Title)
	return nil
}

func (s *Store) GetCourses(ctx context.Context, status *models.CourseStatus) ([]*models.Course, error) {
	query := `
        SELECT id, instructor_id, title, description, status, created_at, updated_at
        FROM courses
        WHERE ($1::text IS NULL OR status = $1)
        ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, status)
	if err != nil {
		log.Printf("failed to query courses: %v", err)
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
			log.Printf("failed to scan course row: %v", err)
			return nil, err
		}
		courses = append(courses, &c)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error iterating course rows: %v", err)
		return nil, err
	}

	log.Printf("retrieved %d courses", len(courses))
	return courses, nil
}

func (s *Store) UpdateCourseStatus(ctx context.Context, courseID int, status models.CourseStatus) error {
	query := `UPDATE courses SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := s.db.Exec(ctx, query, status, courseID)
	if err != nil {
		log.Printf("failed to update course status: courseID=%d, error=%v", courseID, err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no course found with id=%d", courseID)
	} else {
		log.Printf("course status updated: id=%d, status=%s", courseID, status)
	}

	return nil
}

func (s *Store) DeleteCourse(ctx context.Context, courseID int) error {
	query := `DELETE FROM courses WHERE id = $1`
	result, err := s.db.Exec(ctx, query, courseID)
	if err != nil {
		log.Printf("failed to delete course: courseID=%d, error=%v", courseID, err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no course found to delete with id=%d", courseID)
	} else {
		log.Printf("course deleted: id=%d", courseID)
	}

	return nil
}
