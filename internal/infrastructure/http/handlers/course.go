package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/services"
)

type CourseHandler struct {
	Service *services.CourseService
}

func NewCourseHandler(courseService *services.CourseService) *CourseHandler {
	return &CourseHandler{
		Service: courseService,
	}
}

type UpdateCourseRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (r *UpdateCourseRequest) ToCommand(courseID, userID int64) *services.UpdateCourseCommand {
	return &services.UpdateCourseCommand{
		CourseID:             courseID,
		Title:                strings.TrimSpace(r.Title),
		Summary:              strings.TrimSpace(r.Summary),
		TargetAudience:       strings.TrimSpace(r.TargetAudience),
		LearningObjectives:   strings.TrimSpace(r.LearningObjectives),
		AssumedPrerequisites: strings.TrimSpace(r.AssumedPrerequisites),
		UserID:               userID,
	}
}

func (h *CourseHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req UpdateCourseRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.Service.Update(r.Context(), req.ToCommand(courseID, user.ID)); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CourseHandler) Publish(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Publish(r.Context(), &services.PublishCourseCommand{
		CourseID: courseID,
		UserID:   user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CourseHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	course, err := h.Service.Get(r.Context(), &services.GetCourseQuery{
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, course)
}

func (h *CourseHandler) List(w http.ResponseWriter, r *http.Request) {
	courses, err := h.Service.List(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	if courses == nil {
		courses = make([]domain.Course, 0)
	}

	writeJSON(w, http.StatusOK, courses)
}
