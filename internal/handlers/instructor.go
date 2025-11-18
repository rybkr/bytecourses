package handlers

import (
	"encoding/json"
	"github.com/rybkr/bytecourses/internal/middleware"
	"github.com/rybkr/bytecourses/internal/store"
	"log"
	"net/http"
	"strconv"
)

type InstructorHandler struct {
	store *store.Store
}

func NewInstructorHandler(store *store.Store) *InstructorHandler {
	return &InstructorHandler{store: store}
}

func (h *InstructorHandler) GetMyCourses(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	courses, err := h.store.GetCoursesByInstructor(r.Context(), user.ID)
	if err != nil {
		log.Printf("failed to get instructor courses: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

func (h *InstructorHandler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid course id: %s", idStr)
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	course, err := h.store.GetCourseByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to get course: %v", err)
		http.Error(w, "course not found", http.StatusNotFound)
		return
	}

	if course.InstructorID != user.ID {
		log.Printf("user %d attempted to update course %d owned by %d", user.ID, id, course.InstructorID)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var updateData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		log.Printf("failed to decode update request: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateCourse(r.Context(), id, updateData.Title, updateData.Description); err != nil {
		log.Printf("failed to update course: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *InstructorHandler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid course id: %s", idStr)
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	course, err := h.store.GetCourseByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to get course: %v", err)
		http.Error(w, "course not found", http.StatusNotFound)
		return
	}

	if course.InstructorID != user.ID {
		log.Printf("user %d attempted to delete course %d owned by %d", user.ID, id, course.InstructorID)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := h.store.DeleteCourse(r.Context(), id); err != nil {
		log.Printf("failed to delete course: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
