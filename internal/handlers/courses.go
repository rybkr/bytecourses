package handlers

import (
	"encoding/json"
	"github.com/rybkr/bytecourses/internal/models"
	"github.com/rybkr/bytecourses/internal/store"
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
	var course models.Course
	if err := json.NewDecoder(r.Body).Decode(&course); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	course.Status = models.StatusPending

	if err := h.store.CreateCourse(r.Context(), &course); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}

func (h *CourseHandler) ListCourses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.store.GetCourses(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func (h *CourseHandler) ApproveCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateCourseStatus(r.Context(), id, models.StatusApproved); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
