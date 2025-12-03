package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rybkr/bytecourses/internal/helpers"
	"github.com/rybkr/bytecourses/internal/middleware"
	"github.com/rybkr/bytecourses/internal/store"
)

type CourseHandler struct {
	store *store.Store
}

func NewCourseHandler(store *store.Store) *CourseHandler {
	return &CourseHandler{store: store}
}

// CreateCourse removed - moved to applications handler

func (h *CourseHandler) ListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.store.GetCoursesWithInstructors(r.Context())
	if err != nil {
		log.Printf("failed to get courses in handler: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.Success(w, courses)
}

func (h *CourseHandler) GetCourse(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/courses/")
	id, err := strconv.Atoi(path)
	if err != nil {
		log.Printf("invalid course id: %s", path)
		helpers.BadRequest(w, "invalid course id")
		return
	}

	course, instructor, err := h.store.GetCourseWithInstructor(r.Context(), id)
	if err != nil {
		log.Printf("failed to get course: %v", err)
		helpers.NotFound(w, "course not found")
		return
	}

	enrollmentCount, err := h.store.GetEnrollmentCount(r.Context(), id)
	if err != nil {
		log.Printf("failed to get enrollment count: %v", err)
		enrollmentCount = 0
	}

	type EnrollmentData struct {
		IsEnrolled     bool   `json:"is_enrolled"`
		EnrolledAt     string `json:"enrolled_at,omitempty"`
		LastAccessedAt string `json:"last_accessed_at,omitempty"`
	}

	enrollment := EnrollmentData{
		IsEnrolled: false,
	}

	isInstructor := false
	user, ok := middleware.GetUserFromContext(r.Context())
	if ok && user != nil {
		if user.ID == course.InstructorID {
			isInstructor = true
		}

		enrollmentRecord, err := h.store.GetEnrollment(r.Context(), user.ID, id)
		if err == nil && enrollmentRecord != nil {
			enrollment.IsEnrolled = true
			enrollment.EnrolledAt = enrollmentRecord.EnrolledAt.Format("2006-01-02T15:04:05Z07:00")
			if enrollmentRecord.LastAccessedAt != nil {
				enrollment.LastAccessedAt = enrollmentRecord.LastAccessedAt.Format("2006-01-02T15:04:05Z07:00")
			}

			err = h.store.UpdateLastAccessed(r.Context(), user.ID, id)
			if err != nil {
				log.Printf("failed to update last accessed: %v", err)
			}
		}
	}

	type CourseResponse struct {
		*store.CourseWithInstructor
		Enrollment      EnrollmentData `json:"enrollment"`
		EnrollmentCount int            `json:"enrollment_count"`
		IsInstructor    bool           `json:"is_instructor"`
	}

	response := &CourseResponse{
		CourseWithInstructor: &store.CourseWithInstructor{
			Course:          course,
			InstructorName:  instructor.Name,
			InstructorEmail: instructor.Email,
		},
		Enrollment:      enrollment,
		EnrollmentCount: enrollmentCount,
		IsInstructor:    isInstructor,
	}

	helpers.Success(w, response)
}
