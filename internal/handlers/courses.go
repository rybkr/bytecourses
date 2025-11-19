package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rybkr/bytecourses/internal/helpers"
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

	type CourseResponse struct {
		*store.CourseWithInstructor
	}

	response := &CourseResponse{
		CourseWithInstructor: &store.CourseWithInstructor{
			Course:          course,
			InstructorName:  instructor.Name,
			InstructorEmail: instructor.Email,
		},
	}

	helpers.Success(w, response)
}
