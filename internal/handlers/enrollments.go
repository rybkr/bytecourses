package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rybkr/bytecourses/internal/helpers"
	"github.com/rybkr/bytecourses/internal/middleware"
	"github.com/rybkr/bytecourses/internal/models"
	"github.com/rybkr/bytecourses/internal/store"
)

type EnrollmentHandler struct {
	store *store.Store
}

func NewEnrollmentHandler(store *store.Store) *EnrollmentHandler {
	return &EnrollmentHandler{store: store}
}

func (h *EnrollmentHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		helpers.Unauthorized(w, "authentication required")
		return
	}

	if user.Role != models.RoleStudent {
		helpers.Forbidden(w, "only students can enroll in courses")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/courses/")
	path = strings.TrimSuffix(path, "/enroll")
	courseID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("invalid course id: %s", path)
		helpers.BadRequest(w, "invalid course id")
		return
	}

	err = h.store.CreateEnrollment(r.Context(), user.ID, courseID)
	if err != nil {
		log.Printf("failed to create enrollment: %v", err)
		helpers.InternalServerError(w, "failed to enroll in course")
		return
	}

	helpers.Success(w, map[string]interface{}{
		"message": "enrolled successfully",
	})
}

func (h *EnrollmentHandler) Unenroll(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		helpers.Unauthorized(w, "authentication required")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/courses/")
	path = strings.TrimSuffix(path, "/enroll")
	courseID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("invalid course id: %s", path)
		helpers.BadRequest(w, "invalid course id")
		return
	}

	err = h.store.DeleteEnrollment(r.Context(), user.ID, courseID)
	if err != nil {
		log.Printf("failed to delete enrollment: %v", err)
		helpers.InternalServerError(w, "failed to unenroll from course")
		return
	}

	helpers.NoContent(w)
}

func (h *EnrollmentHandler) GetEnrollmentStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		helpers.Success(w, map[string]interface{}{
			"is_enrolled": false,
		})
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/courses/")
	path = strings.TrimSuffix(path, "/enrollments")
	courseID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("invalid course id: %s", path)
		helpers.BadRequest(w, "invalid course id")
		return
	}

	enrollment, err := h.store.GetEnrollment(r.Context(), user.ID, courseID)
	if err != nil {
		helpers.Success(w, map[string]interface{}{
			"is_enrolled": false,
		})
		return
	}

	enrollmentCount, _ := h.store.GetEnrollmentCount(r.Context(), courseID)

	helpers.Success(w, map[string]interface{}{
		"is_enrolled":      true,
		"enrolled_at":      enrollment.EnrolledAt,
		"last_accessed_at": enrollment.LastAccessedAt,
		"enrollment_count": enrollmentCount,
	})
}

func (h *EnrollmentHandler) ListEnrollments(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		helpers.Unauthorized(w, "authentication required")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/courses/")
	path = strings.TrimSuffix(path, "/enrollments")
	courseID, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("invalid course id: %s", path)
		helpers.BadRequest(w, "invalid course id")
		return
	}

	course, _, err := h.store.GetCourseWithInstructor(r.Context(), courseID)
	if err != nil {
		log.Printf("failed to get course: %v", err)
		helpers.NotFound(w, "course not found")
		return
	}

	if user.Role != models.RoleAdmin && user.ID != course.InstructorID {
		helpers.Forbidden(w, "only instructor or admin can view enrollment list")
		return
	}

	enrollments, err := h.store.GetEnrollmentsByCourse(r.Context(), courseID)
	if err != nil {
		log.Printf("failed to get enrollments: %v", err)
		helpers.InternalServerError(w, "failed to get enrollments")
		return
	}

	helpers.Success(w, enrollments)
}

func (h *EnrollmentHandler) GetStudentEnrollments(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok || user == nil {
		helpers.Unauthorized(w, "authentication required")
		return
	}

	enrollments, err := h.store.GetEnrollmentsByStudent(r.Context(), user.ID)
	if err != nil {
		log.Printf("failed to get student enrollments: %v", err)
		helpers.InternalServerError(w, "failed to get enrollments")
		return
	}

	helpers.Success(w, enrollments)
}
