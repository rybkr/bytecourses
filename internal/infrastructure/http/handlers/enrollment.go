package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/services"
)

type EnrollmentHandler struct {
	Service *services.EnrollmentService
}

func NewEnrollmentHandler(enrollmentService *services.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{
		Service: enrollmentService,
	}
}

func (h *EnrollmentHandler) Enroll(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	if err := h.Service.Enroll(r.Context(), &services.EnrollCommand{
		CourseID: courseID,
		UserID:   user.ID,
	}); err != nil {
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EnrollmentHandler) Unenroll(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	if err := h.Service.Unenroll(r.Context(), &services.UnenrollCommand{
		CourseID: courseID,
		UserID:   user.ID,
	}); err != nil {
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *EnrollmentHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	isEnrolled, err := h.Service.IsEnrolled(r.Context(), &services.IsEnrolledQuery{
		CourseID: courseID,
		UserID:   user.ID,
	})
	if err != nil {
		handleError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{
		"enrolled": isEnrolled,
	})
}

func (h *EnrollmentHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	enrollments, err := h.Service.ListByUser(r.Context(), &services.ListEnrollmentsByUserQuery{
		UserID: user.ID,
	})
	if err != nil {
		handleError(w, r, err)
		return
	}

	if enrollments == nil {
		enrollments = make([]domain.Enrollment, 0)
	}

	writeJSON(w, http.StatusOK, enrollments)
}
