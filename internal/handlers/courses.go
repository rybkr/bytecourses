package handlers

import (
	"encoding/json"
	"github.com/rybkr/bytecourses/internal/middleware"
	"github.com/rybkr/bytecourses/internal/models"
	"github.com/rybkr/bytecourses/internal/store"
	"log"
	"net/http"
	"strconv"
)

type CourseHandler struct {
	store *store.Store
}

func NewCourseHandler(store *store.Store) *CourseHandler {
	return &CourseHandler{store: store}
}

func (h *CourseHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		log.Printf("failed to decode course request: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	course.InstructorID = user.ID
	course.Status = models.StatusPending

	if err := h.store.CreateCourse(r.Context(), &course); err != nil {
		log.Printf("failed to create course in handler: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(course); err != nil {
		log.Printf("failed to encode course response: %v", err)
	}
}

func (h *CourseHandler) ListCourses(w http.ResponseWriter, r *http.Request) {
	var status *models.CourseStatus
	if s := r.URL.Query().Get("status"); s != "" {
		st := models.CourseStatus(s)
		status = &st
	}

	courses, err := h.store.GetCourses(r.Context(), status)
	if err != nil {
		log.Printf("failed to get courses in handler: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(courses); err != nil {
		log.Printf("failed to encode courses response: %v", err)
	}
}

func (h *CourseHandler) ApproveCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid course id: %s", idStr)
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateCourseStatus(r.Context(), id, models.StatusApproved); err != nil {
		log.Printf("failed to approve course in handler: id=%d, error=%v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *CourseHandler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid course id for deletion: %s", idStr)
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteCourse(r.Context(), id); err != nil {
		log.Printf("failed to delete course in handler: id=%d, error=%v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
