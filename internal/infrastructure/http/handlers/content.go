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

type CreateContentRequest struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Order   int    `json:"order"`
	Format  string `json:"format"`
	Content string `json:"content"`
}

func (r *CreateContentRequest) ToCommand(moduleID, userID int64) *services.CreateContentCommand {
	return &services.CreateContentCommand{
		Type:     domain.ContentType(strings.TrimSpace(r.Type)),
		ModuleID: moduleID,
		Title:    strings.TrimSpace(r.Title),
		Order:    r.Order,
		Format:   strings.TrimSpace(r.Format),
		Content:  strings.TrimSpace(r.Content),
		UserID:   userID,
	}
}

func (h *ContentHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req CreateContentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	content, err := h.Service.Create(r.Context(), req.ToCommand(moduleID, user.ID))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, content)
}

type UpdateContentRequest struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Order   int    `json:"order"`
	Format  string `json:"format"`
	Content string `json:"content"`
}

func (r *UpdateContentRequest) ToCommand(contentID, userID int64) *services.UpdateContentCommand {
	return &services.UpdateContentCommand{
		Type:      domain.ContentType(strings.TrimSpace(r.Type)),
		ContentID: contentID,
		Title:     strings.TrimSpace(r.Title),
		Order:     r.Order,
		Format:    strings.TrimSpace(r.Format),
		Content:   strings.TrimSpace(r.Content),
		UserID:    userID,
	}
}

func (h *ContentHandler) Update(w http.ResponseWriter, r *http.Request) {
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

	contentID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	var req UpdateContentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.Service.Update(r.Context(), req.ToCommand(contentID, user.ID)); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

	contentID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	contentTypeStr := r.URL.Query().Get("type")
	if contentTypeStr == "" {
		contentTypeStr = string(domain.ContentTypeReading)
	}

	if err := h.Service.Delete(r.Context(), &services.DeleteContentCommand{
		Type:      domain.ContentType(contentTypeStr),
		ContentID: contentID,
		UserID:    user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) Publish(w http.ResponseWriter, r *http.Request) {
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

	contentID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	contentTypeStr := r.URL.Query().Get("type")
	if contentTypeStr == "" {
		contentTypeStr = string(domain.ContentTypeReading)
	}

	if err := h.Service.Publish(r.Context(), &services.PublishContentCommand{
		Type:      domain.ContentType(contentTypeStr),
		ContentID: contentID,
		UserID:    user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
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

	contentID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	contentTypeStr := r.URL.Query().Get("type")
	if contentTypeStr == "" {
		contentTypeStr = string(domain.ContentTypeReading)
	}

	if err := h.Service.Unpublish(r.Context(), &services.UnpublishContentCommand{
		Type:      domain.ContentType(contentTypeStr),
		ContentID: contentID,
		UserID:    user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) List(w http.ResponseWriter, r *http.Request) {
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

	items, err := h.Service.List(r.Context(), &services.ListContentQuery{
		ModuleID: moduleID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		handleError(w, err)
		return
	}
	if items == nil {
		items = make([]domain.ContentItem, 0)
	}

	writeJSON(w, http.StatusOK, items)
}

func (h *ContentHandler) Get(w http.ResponseWriter, r *http.Request) {
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

	contentID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid content id", http.StatusBadRequest)
		return
	}

	content, err := h.Service.Get(r.Context(), &services.GetContentQuery{
		ContentID: contentID,
		ModuleID:  moduleID,
		UserID:    user.ID,
		UserRole:  user.Role,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, content)
}
