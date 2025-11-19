package store

import (
	"context"
	"log"

	"github.com/rybkr/bytecourses/internal/models"
)

func (s *Store) CreateCourse(ctx context.Context, course *models.Course) error {
	query := `
        INSERT INTO courses (instructor_id, title, description, content)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query,
		course.InstructorID,
		course.Title,
		course.Description,
		course.Content,
	).Scan(&course.ID, &course.CreatedAt, &course.UpdatedAt)

	if err != nil {
		log.Printf("failed to create course: %v", err)
		return err
	}

	log.Printf("course created: id=%d, title=%s", course.ID, course.Title)
	return nil
}

func (s *Store) CreateCourseFromApplication(ctx context.Context, app *models.Application) (*models.Course, error) {
	course := &models.Course{
		InstructorID: app.InstructorID,
		Title:        app.Title,
		Description:  app.Description,
	}

	err := s.CreateCourse(ctx, course)
	if err != nil {
		return nil, err
	}

	return course, nil
}

func (s *Store) GetCourses(ctx context.Context) ([]*models.Course, error) {
	query := `
        SELECT id, instructor_id, title, description, content, created_at, updated_at
        FROM courses
        ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		log.Printf("failed to query courses: %v", err)
		return nil, err
	}
	defer rows.Close()

	var courses []*models.Course
	for rows.Next() {
		var c models.Course
		err := rows.Scan(
			&c.ID, &c.InstructorID, &c.Title, &c.Description, &c.Content,
			&c.CreatedAt, &c.UpdatedAt,
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

// UpdateCourseStatus removed - status management moved to applications

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
        SELECT c.id, c.instructor_id, c.title, c.description, c.content, c.created_at, c.updated_at,
               u.id, u.email, u.role, u.created_at
        FROM courses c
        JOIN users u ON c.instructor_id = u.id
        WHERE c.id = $1`

	var course models.Course
	var instructor models.User

	err := s.db.QueryRow(ctx, query, courseID).Scan(
		&course.ID, &course.InstructorID, &course.Title, &course.Description, &course.Content,
		&course.CreatedAt, &course.UpdatedAt,
		&instructor.ID, &instructor.Email, &instructor.Role, &instructor.CreatedAt,
	)

	if err != nil {
		log.Printf("failed to get course with instructor: %v", err)
		return nil, nil, err
	}

	return &course, &instructor, nil
}

// RejectCourse removed - rejection moved to applications

func (s *Store) GetCoursesByInstructor(ctx context.Context, instructorID int) ([]*models.Course, error) {
	query := `
        SELECT id, instructor_id, title, description, created_at, updated_at
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
			&c.CreatedAt, &c.UpdatedAt,
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

// CountDraftsByInstructor moved to applications.go

func (s *Store) GetCourseByID(ctx context.Context, courseID int) (*models.Course, error) {
	var course models.Course
	query := `
        SELECT id, instructor_id, title, description, created_at, updated_at
        FROM courses
        WHERE id = $1`

	err := s.db.QueryRow(ctx, query, courseID).Scan(
		&course.ID, &course.InstructorID, &course.Title, &course.Description,
		&course.CreatedAt, &course.UpdatedAt,
	)

	if err != nil {
		log.Printf("failed to get course by id: %v", err)
		return nil, err
	}

	return &course, nil
}

type CourseWithInstructor struct {
	*models.Course
	InstructorName  string `json:"instructor_name"`
	InstructorEmail string `json:"instructor_email"`
}

func (s *Store) GetCoursesWithInstructors(ctx context.Context) ([]*CourseWithInstructor, error) {
	query := `
        SELECT c.id, c.instructor_id, c.title, c.description, c.created_at, c.updated_at,
               COALESCE(u.name, ''), u.email
        FROM courses c
        JOIN users u ON c.instructor_id = u.id
        ORDER BY c.created_at DESC`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		log.Printf("failed to query courses with instructors: %v", err)
		return nil, err
	}
	defer rows.Close()

	var courses []*CourseWithInstructor
	for rows.Next() {
		var cwi CourseWithInstructor
		var c models.Course
		err := rows.Scan(
			&c.ID, &c.InstructorID, &c.Title, &c.Description,
			&c.CreatedAt, &c.UpdatedAt,
			&cwi.InstructorName, &cwi.InstructorEmail,
		)
		if err != nil {
			log.Printf("failed to scan course row: %v", err)
			return nil, err
		}
		cwi.Course = &c
		courses = append(courses, &cwi)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error iterating course rows: %v", err)
		return nil, err
	}

	log.Printf("retrieved %d courses with instructors", len(courses))
	return courses, nil
}
