package store

import (
	"context"
	"fmt"
	"log"

	"github.com/rybkr/bytecourses/internal/models"
)

func (s *Store) CreateApplication(ctx context.Context, app *models.Application) error {
	query := `
        INSERT INTO applications (instructor_id, title, description, learning_objectives, prerequisites, 
                                  course_format, category_tags, skill_level, course_duration, 
                                  instructor_qualifications, status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query,
		app.InstructorID,
		app.Title,
		app.Description,
		app.LearningObjectives,
		app.Prerequisites,
		app.CourseFormat,
		app.CategoryTags,
		app.SkillLevel,
		app.CourseDuration,
		app.InstructorQualifications,
		app.Status,
	).Scan(&app.ID, &app.CreatedAt, &app.UpdatedAt)

	if err != nil {
		log.Printf("failed to create application: %v", err)
		return err
	}
	log.Printf("application created: id=%d, title=%s", app.ID, app.Title)
	return nil
}

func (s *Store) GetApplicationByID(ctx context.Context, id int) (*models.Application, error) {
	query := `
        SELECT id, instructor_id, title, description, learning_objectives, prerequisites,
               course_format, category_tags, skill_level, course_duration, instructor_qualifications,
               status, rejected_at, created_at, updated_at
        FROM applications
        WHERE id = $1`

	var app models.Application
	err := s.db.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.InstructorID, &app.Title, &app.Description,
		&app.LearningObjectives, &app.Prerequisites, &app.CourseFormat,
		&app.CategoryTags, &app.SkillLevel, &app.CourseDuration,
		&app.InstructorQualifications, &app.Status, &app.RejectedAt,
		&app.CreatedAt, &app.UpdatedAt,
	)

	if err != nil {
		log.Printf("failed to get application: %v", err)
		return nil, err
	}

	return &app, nil
}

func (s *Store) GetApplicationsByInstructor(ctx context.Context, instructorID int) ([]*models.Application, error) {
	query := `
        SELECT id, instructor_id, title, description, learning_objectives, prerequisites,
               course_format, category_tags, skill_level, course_duration, instructor_qualifications,
               status, rejected_at, created_at, updated_at
        FROM applications
        WHERE instructor_id = $1
        ORDER BY created_at DESC`

	rows, err := s.db.Query(ctx, query, instructorID)
	if err != nil {
		log.Printf("failed to query applications by instructor: %v", err)
		return nil, err
	}
	defer rows.Close()

	var applications []*models.Application
	for rows.Next() {
		var app models.Application
		err := rows.Scan(
			&app.ID, &app.InstructorID, &app.Title, &app.Description,
			&app.LearningObjectives, &app.Prerequisites, &app.CourseFormat,
			&app.CategoryTags, &app.SkillLevel, &app.CourseDuration,
			&app.InstructorQualifications, &app.Status, &app.RejectedAt,
			&app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			log.Printf("failed to scan application row: %v", err)
			return nil, err
		}
		applications = append(applications, &app)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error iterating application rows: %v", err)
		return nil, err
	}

	log.Printf("retrieved %d applications for instructor %d", len(applications), instructorID)
	return applications, nil
}

func (s *Store) GetPendingApplications(ctx context.Context) ([]*models.Application, error) {
	query := `
        SELECT a.id, a.instructor_id, a.title, a.description, a.learning_objectives, a.prerequisites,
               a.course_format, a.category_tags, a.skill_level, a.course_duration, a.instructor_qualifications,
               a.status, a.rejected_at, a.created_at, a.updated_at,
               u.id, u.email, u.role, u.created_at
        FROM applications a
        JOIN users u ON a.instructor_id = u.id
        WHERE a.status = 'pending'
        ORDER BY a.created_at DESC`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		log.Printf("failed to query pending applications: %v", err)
		return nil, err
	}
	defer rows.Close()

	var applications []*models.Application
	for rows.Next() {
		var app models.Application
		var userID int
		var userEmail, userRole string
		var userCreatedAt interface{}
		err := rows.Scan(
			&app.ID, &app.InstructorID, &app.Title, &app.Description,
			&app.LearningObjectives, &app.Prerequisites, &app.CourseFormat,
			&app.CategoryTags, &app.SkillLevel, &app.CourseDuration,
			&app.InstructorQualifications, &app.Status, &app.RejectedAt,
			&app.CreatedAt, &app.UpdatedAt,
			&userID, &userEmail, &userRole, &userCreatedAt,
		)
		if err != nil {
			log.Printf("failed to scan pending application row: %v", err)
			return nil, err
		}
		applications = append(applications, &app)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error iterating pending application rows: %v", err)
		return nil, err
	}

	log.Printf("retrieved %d pending applications", len(applications))
	return applications, nil
}

func (s *Store) UpdateApplication(ctx context.Context, id int, updates map[string]interface{}) error {
	// Build dynamic UPDATE query based on provided fields
	setParts := []string{}
	args := []interface{}{}
	argPos := 1

	allowedFields := map[string]bool{
		"title": true, "description": true, "learning_objectives": true,
		"prerequisites": true, "course_format": true, "category_tags": true,
		"skill_level": true, "course_duration": true, "instructor_qualifications": true,
		"status": true,
	}

	for field, value := range updates {
		if allowedFields[field] {
			setParts = append(setParts, field+" = $"+fmt.Sprintf("%d", argPos))
			args = append(args, value)
			argPos++
		}
	}

	if len(setParts) == 0 {
		return nil
	}

	args = append(args, id)
	query := "UPDATE applications SET " + setParts[0]
	for i := 1; i < len(setParts); i++ {
		query += ", " + setParts[i]
	}
	query += ", updated_at = NOW() WHERE id = $" + fmt.Sprintf("%d", argPos)

	result, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		log.Printf("failed to update application: applicationID=%d, error=%v", id, err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no application found with id=%d", id)
	} else {
		log.Printf("application updated: id=%d", id)
	}
	return nil
}

func (s *Store) DeleteApplication(ctx context.Context, id int) error {
	query := `DELETE FROM applications WHERE id = $1`
	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		log.Printf("failed to delete application: applicationID=%d, error=%v", id, err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no application found to delete with id=%d", id)
	} else {
		log.Printf("application deleted: id=%d", id)
	}

	return nil
}

func (s *Store) CountDraftsByInstructor(ctx context.Context, instructorID int) (int, error) {
	query := `SELECT COUNT(*) FROM applications WHERE instructor_id = $1 AND status = 'draft'`
	var count int
	err := s.db.QueryRow(ctx, query, instructorID).Scan(&count)
	if err != nil {
		log.Printf("failed to count drafts: instructorID=%d, error=%v", instructorID, err)
		return 0, err
	}
	return count, nil
}

func (s *Store) ApproveApplication(ctx context.Context, id int) (*models.Application, error) {
	app, err := s.GetApplicationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if app.Status != models.StatusPending {
		log.Printf("cannot approve application with status %s", app.Status)
		return nil, err
	}

	return app, nil
}

func (s *Store) RejectApplication(ctx context.Context, id int) error {
	query := `UPDATE applications SET status = 'rejected', rejected_at = NOW(), updated_at = NOW() WHERE id = $1`
	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		log.Printf("failed to reject application: applicationID=%d, error=%v", id, err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("no application found with id=%d", id)
	} else {
		log.Printf("application rejected: id=%d", id)
	}

	return nil
}

func (s *Store) DeleteExpiredRejectedApplications(ctx context.Context, ttlDays int) error {
	query := `DELETE FROM applications WHERE status = 'rejected' AND rejected_at < NOW() - INTERVAL '1 day' * $1`
	result, err := s.db.Exec(ctx, query, ttlDays)
	if err != nil {
		log.Printf("failed to delete expired rejected applications: error=%v", err)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("deleted %d expired rejected applications (TTL: %d days)", rowsAffected, ttlDays)
	}

	return nil
}

func (s *Store) GetExpiredRejectedApplications(ctx context.Context, ttlDays int) ([]*models.Application, error) {
	query := `
        SELECT id, instructor_id, title, description, learning_objectives, prerequisites,
               course_format, category_tags, skill_level, course_duration, instructor_qualifications,
               status, rejected_at, created_at, updated_at
        FROM applications
        WHERE status = 'rejected' AND rejected_at < NOW() - INTERVAL '1 day' * $1`

	rows, err := s.db.Query(ctx, query, ttlDays)
	if err != nil {
		log.Printf("failed to query expired rejected applications: %v", err)
		return nil, err
	}
	defer rows.Close()

	var applications []*models.Application
	for rows.Next() {
		var app models.Application
		err := rows.Scan(
			&app.ID, &app.InstructorID, &app.Title, &app.Description,
			&app.LearningObjectives, &app.Prerequisites, &app.CourseFormat,
			&app.CategoryTags, &app.SkillLevel, &app.CourseDuration,
			&app.InstructorQualifications, &app.Status, &app.RejectedAt,
			&app.CreatedAt, &app.UpdatedAt,
		)
		if err != nil {
			log.Printf("failed to scan expired application row: %v", err)
			return nil, err
		}
		applications = append(applications, &app)
	}

	return applications, nil
}

