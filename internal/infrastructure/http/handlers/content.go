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

type ContentHandler struct {
	Service *services.ContentService
}

func NewContentHandler(contentService *services.ContentService) *ContentHandler {
	return &ContentHandler{
		Service: contentService,
	}
}

type CreateReadingRequest struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Order   int    `json:"order"`
	Format  string `json:"format"`
	Content string `json:"content"`
}

func (r *CreateReadingRequest) ToCommand(moduleID, userID int64) *services.CreateReadingCommand {
	return &services.CreateReadingCommand{
		ModuleID: moduleID,
		Title:    strings.TrimSpace(r.Title),
		Order:    r.Order,
		Format:   strings.TrimSpace(r.Format),
		Content:  strings.TrimSpace(r.Content),
		UserID:   userID,
	}
}

func (h *ContentHandler) CreateReading(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	var req CreateReadingRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Type != string(domain.ContentTypeReading) {
		http.Error(w, "invalid content type", http.StatusBadRequest)
		return
	}

	reading, err := h.Service.CreateReading(r.Context(), req.ToCommand(moduleID, user.ID))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, reading)
}

type UpdateReadingRequest struct {
	Title   string `json:"title"`
	Order   int    `json:"order"`
	Format  string `json:"format"`
	Content string `json:"content"`
}

func (r *UpdateReadingRequest) ToCommand(readingID, userID int64) *services.UpdateReadingCommand {
	return &services.UpdateReadingCommand{
		ReadingID: readingID,
		Title:     strings.TrimSpace(r.Title),
		Order:     r.Order,
		Format:    strings.TrimSpace(r.Format),
		Content:   strings.TrimSpace(r.Content),
		UserID:    userID,
	}
}

func (h *ContentHandler) UpdateReading(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	_, err = strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	readingID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	var req UpdateReadingRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.Service.UpdateReading(r.Context(), req.ToCommand(readingID, user.ID)); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) DeleteReading(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	_, err = strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	readingID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteReading(r.Context(), &services.DeleteReadingCommand{
		ReadingID: readingID,
		UserID:    user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) PublishReading(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	_, err = strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	readingID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	if err := h.Service.PublishReading(r.Context(), &services.PublishReadingCommand{
		ReadingID: readingID,
		UserID:    user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) ListReadings(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	readings, err := h.Service.ListReadings(r.Context(), &services.ListReadingsQuery{
		ModuleID: moduleID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		handleError(w, err)
		return
	}
	if readings == nil {
		readings = make([]domain.Reading, 0)
	}

	writeJSON(w, http.StatusOK, readings)
}

func (h *ContentHandler) GetReading(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	readingID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	reading, err := h.Service.GetReading(r.Context(), &services.GetReadingQuery{
		ReadingID: readingID,
		ModuleID:  moduleID,
		UserID:    user.ID,
		UserRole:  user.Role,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, reading)
}
