package store

import (
	"context"
	"log"
	"time"

	"github.com/rybkr/bytecourses/internal/models"
)

func (s *Store) CreateEnrollment(ctx context.Context, studentID, courseID int) error {
	query := `
        INSERT INTO enrollments (student_id, course_id, enrolled_at)
        VALUES ($1, $2, NOW())
        ON CONFLICT (student_id, course_id) DO NOTHING
        RETURNING id, enrolled_at`

	var id int
	var enrolledAt time.Time
	err := s.db.QueryRow(ctx, query, studentID, courseID).Scan(&id, &enrolledAt)

	if err != nil {
		if err.Error() == "no rows in result set" {
			log.Printf("enrollment already exists: student_id=%d, course_id=%d", studentID, courseID)
			return nil
		}
		log.Printf("failed to create enrollment: %v", err)
		return err
	}

	log.Printf("enrollment created: id=%d, student_id=%d, course_id=%d", id, studentID, courseID)
	return nil
}

func (s *Store) DeleteEnrollment(ctx context.Context, studentID, courseID int) error {
	query := `DELETE FROM enrollments WHERE student_id = $1 AND course_id = $2`

	result, err := s.db.Exec(ctx, query, studentID, courseID)
	if err != nil {
		log.Printf("failed to delete enrollment: %v", err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no enrollment found: student_id=%d, course_id=%d", studentID, courseID)
	} else {
		log.Printf("enrollment deleted: student_id=%d, course_id=%d", studentID, courseID)
	}

	return nil
}

func (s *Store) GetEnrollment(ctx context.Context, studentID, courseID int) (*models.Enrollment, error) {
	query := `
        SELECT id, student_id, course_id, enrolled_at, last_accessed_at
        FROM enrollments
        WHERE student_id = $1 AND course_id = $2`

	var enrollment models.Enrollment
	var lastAccessedAt *time.Time

	err := s.db.QueryRow(ctx, query, studentID, courseID).Scan(
		&enrollment.ID,
		&enrollment.StudentID,
		&enrollment.CourseID,
		&enrollment.EnrolledAt,
		&lastAccessedAt,
	)

	if err != nil {
		return nil, err
	}

	enrollment.LastAccessedAt = lastAccessedAt
	return &enrollment, nil
}

func (s *Store) GetEnrollmentsByStudent(ctx context.Context, studentID int) ([]*models.Enrollment, error) {
	query := `
        SELECT id, student_id, course_id, enrolled_at, last_accessed_at
        FROM enrollments
        WHERE student_id = $1
        ORDER BY enrolled_at DESC`

	rows, err := s.db.Query(ctx, query, studentID)
	if err != nil {
		log.Printf("failed to query enrollments by student: %v", err)
		return nil, err
	}
	defer rows.Close()

	var enrollments []*models.Enrollment
	for rows.Next() {
		var enrollment models.Enrollment
		var lastAccessedAt *time.Time

		err := rows.Scan(
			&enrollment.ID,
			&enrollment.StudentID,
			&enrollment.CourseID,
			&enrollment.EnrolledAt,
			&lastAccessedAt,
		)
		if err != nil {
			log.Printf("failed to scan enrollment row: %v", err)
			return nil, err
		}

		enrollment.LastAccessedAt = lastAccessedAt
		enrollments = append(enrollments, &enrollment)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error iterating enrollment rows: %v", err)
		return nil, err
	}

	return enrollments, nil
}

func (s *Store) GetEnrollmentsByCourse(ctx context.Context, courseID int) ([]*models.Enrollment, error) {
	query := `
        SELECT id, student_id, course_id, enrolled_at, last_accessed_at
        FROM enrollments
        WHERE course_id = $1
        ORDER BY enrolled_at DESC`

	rows, err := s.db.Query(ctx, query, courseID)
	if err != nil {
		log.Printf("failed to query enrollments by course: %v", err)
		return nil, err
	}
	defer rows.Close()

	var enrollments []*models.Enrollment
	for rows.Next() {
		var enrollment models.Enrollment
		var lastAccessedAt *time.Time

		err := rows.Scan(
			&enrollment.ID,
			&enrollment.StudentID,
			&enrollment.CourseID,
			&enrollment.EnrolledAt,
			&lastAccessedAt,
		)
		if err != nil {
			log.Printf("failed to scan enrollment row: %v", err)
			return nil, err
		}

		enrollment.LastAccessedAt = lastAccessedAt
		enrollments = append(enrollments, &enrollment)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error iterating enrollment rows: %v", err)
		return nil, err
	}

	return enrollments, nil
}

func (s *Store) UpdateLastAccessed(ctx context.Context, studentID, courseID int) error {
	query := `
        UPDATE enrollments
        SET last_accessed_at = NOW()
        WHERE student_id = $1 AND course_id = $2`

	_, err := s.db.Exec(ctx, query, studentID, courseID)
	if err != nil {
		log.Printf("failed to update last accessed: %v", err)
		return err
	}

	return nil
}

func (s *Store) GetEnrollmentCount(ctx context.Context, courseID int) (int, error) {
	query := `SELECT COUNT(*) FROM enrollments WHERE course_id = $1`

	var count int
	err := s.db.QueryRow(ctx, query, courseID).Scan(&count)
	if err != nil {
		log.Printf("failed to get enrollment count: %v", err)
		return 0, err
	}

	return count, nil
}

