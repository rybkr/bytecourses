package handlers

import (
	"log"
	"net/http"

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
