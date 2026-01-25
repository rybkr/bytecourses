package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
)

var (
	_ persistence.EnrollmentRepository = (*EnrollmentRepository)(nil)
)

type EnrollmentRepository struct {
	db *sql.DB
}

func NewEnrollmentRepository(db *DB) *EnrollmentRepository {
	return &EnrollmentRepository{
		db: db.DB(),
	}
}

func (r *EnrollmentRepository) Create(ctx context.Context, e *domain.Enrollment) error {
	enrolledAt := time.Now().UTC()

	err := r.db.QueryRowContext(ctx, `
		INSERT INTO enrollments (user_id, course_id, enrolled_at)
		VALUES ($1, $2, $3)
		RETURNING enrolled_at
	`, e.UserID, e.CourseID, enrolledAt).Scan(&e.EnrolledAt)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return errors.ErrConflict
		}
		return err
	}

	return nil
}

func (r *EnrollmentRepository) GetByUserAndCourse(ctx context.Context, userID, courseID int64) (*domain.Enrollment, bool) {
	var e domain.Enrollment

	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, course_id, enrolled_at
		FROM enrollments
		WHERE user_id = $1 AND course_id = $2
	`, userID, courseID).Scan(
		&e.UserID,
		&e.CourseID,
		&e.EnrolledAt,
	)

	if err != nil {
		return nil, false
	}

	return &e, true
}

func (r *EnrollmentRepository) ListByUser(ctx context.Context, userID int64) ([]domain.Enrollment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, course_id, enrolled_at
		FROM enrollments
		WHERE user_id = $1
		ORDER BY enrolled_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	enrollments := make([]domain.Enrollment, 0)
	for rows.Next() {
		var e domain.Enrollment
		if err := rows.Scan(
			&e.UserID,
			&e.CourseID,
			&e.EnrolledAt,
		); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, e)
	}

	return enrollments, rows.Err()
}

func (r *EnrollmentRepository) ListByCourse(ctx context.Context, courseID int64) ([]domain.Enrollment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, course_id, enrolled_at
		FROM enrollments
		WHERE course_id = $1
		ORDER BY enrolled_at DESC
	`, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	enrollments := make([]domain.Enrollment, 0)
	for rows.Next() {
		var e domain.Enrollment
		if err := rows.Scan(
			&e.UserID,
			&e.CourseID,
			&e.EnrolledAt,
		); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, e)
	}

	return enrollments, rows.Err()
}

func (r *EnrollmentRepository) Delete(ctx context.Context, userID, courseID int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM enrollments
		WHERE user_id = $1 AND course_id = $2
	`, userID, courseID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound
	}

	return nil
}
