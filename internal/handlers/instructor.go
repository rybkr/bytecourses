package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/rybkr/bytecourses/internal/helpers"
	"github.com/rybkr/bytecourses/internal/middleware"
	"github.com/rybkr/bytecourses/internal/store"
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
		helpers.Unauthorized(w, "unauthorized")
		return
	}

	courses, err := h.store.GetCoursesByInstructor(r.Context(), user.ID)
	if err != nil {
		log.Printf("failed to get instructor courses: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.Success(w, courses)
}

func (h *InstructorHandler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		helpers.Unauthorized(w, "unauthorized")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid course id: %s", idStr)
		helpers.BadRequest(w, "invalid course id")
		return
	}

	course, err := h.store.GetCourseByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to get course: %v", err)
		helpers.NotFound(w, "course not found")
		return
	}

	if course.InstructorID != user.ID {
		log.Printf("user %d attempted to update course %d owned by %d", user.ID, id, course.InstructorID)
		helpers.Forbidden(w, "forbidden")
		return
	}

	var updateData struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Content     string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		log.Printf("failed to decode update request: %v", err)
		helpers.BadRequest(w, "invalid request body")
		return
	}

	if err := h.store.UpdateCourse(r.Context(), id, updateData.Title, updateData.Description, updateData.Content); err != nil {
		log.Printf("failed to update course: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.NoContent(w)
}

func (h *InstructorHandler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		helpers.Unauthorized(w, "unauthorized")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid course id: %s", idStr)
		helpers.BadRequest(w, "invalid course id")
		return
	}

	course, err := h.store.GetCourseByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to get course: %v", err)
		helpers.NotFound(w, "course not found")
		return
	}

	if course.InstructorID != user.ID {
		log.Printf("user %d attempted to delete course %d owned by %d", user.ID, id, course.InstructorID)
		helpers.Forbidden(w, "forbidden")
		return
	}

	if err := h.store.DeleteCourse(r.Context(), id); err != nil {
		log.Printf("failed to delete course: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.NoContent(w)
}
