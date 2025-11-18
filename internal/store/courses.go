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

func (s *Store) GetCourseWithInstructor(ctx context.Context, courseID int) (*models.Course, *models.User, error) {
	query := `
        SELECT c.id, c.instructor_id, c.title, c.description, c.status, c.created_at, c.updated_at,
               u.id, u.email, u.role, u.created_at
        FROM courses c
        JOIN users u ON c.instructor_id = u.id
        WHERE c.id = $1`

	var course models.Course
	var instructor models.User

	err := s.db.QueryRow(ctx, query, courseID).Scan(
		&course.ID, &course.InstructorID, &course.Title, &course.Description,
		&course.Status, &course.CreatedAt, &course.UpdatedAt,
		&instructor.ID, &instructor.Email, &instructor.Role, &instructor.CreatedAt,
	)

	if err != nil {
		log.Printf("failed to get course with instructor: %v", err)
		return nil, nil, err
	}

	return &course, &instructor, nil
}

func (s *Store) RejectCourse(ctx context.Context, courseID int) error {
	query := `UPDATE courses SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := s.db.Exec(ctx, query, models.StatusRejected, courseID)
	if err != nil {
		log.Printf("failed to reject course: courseID=%d, error=%v", courseID, err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no course found with id=%d", courseID)
	} else {
		log.Printf("course rejected: id=%d", courseID)
	}

	return nil
}

func (s *Store) GetCoursesByInstructor(ctx context.Context, instructorID int) ([]*models.Course, error) {
	query := `
        SELECT id, instructor_id, title, description, status, created_at, updated_at
        FROM courses
        WHERE instructor_id = $1
        ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, instructorID)
	if err != nil {
		log.Printf("failed to query courses by instructor: %v", err)
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

	log.Printf("retrieved %d courses for instructor %d", len(courses), instructorID)
	return courses, nil
}

func (s *Store) UpdateCourse(ctx context.Context, courseID int, title, description string) error {
	query := `
        UPDATE courses 
        SET title = $1, description = $2, updated_at = NOW() 
        WHERE id = $3`

	result, err := s.db.Exec(ctx, query, title, description, courseID)
	if err != nil {
		log.Printf("failed to update course: courseID=%d, error=%v", courseID, err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no course found with id=%d", courseID)
	} else {
		log.Printf("course updated: id=%d", courseID)
	}

	return nil
}

func (s *Store) GetCourseByID(ctx context.Context, courseID int) (*models.Course, error) {
	var course models.Course
	query := `
        SELECT id, instructor_id, title, description, status, created_at, updated_at
        FROM courses
        WHERE id = $1`

	err := s.db.QueryRow(ctx, query, courseID).Scan(
		&course.ID, &course.InstructorID, &course.Title, &course.Description,
		&course.Status, &course.CreatedAt, &course.UpdatedAt,
	)

	if err != nil {
		log.Printf("failed to get course by id: %v", err)
		return nil, err
	}

	return &course, nil
}
