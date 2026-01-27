package handlers

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/services"
)

type ContentHandler struct {
	Service           *services.ContentService
	EnrollmentService *services.EnrollmentService
	CourseService     *services.CourseService
}

func NewContentHandler(contentService *services.ContentService, enrollmentService *services.EnrollmentService, courseService *services.CourseService) *ContentHandler {
	return &ContentHandler{
		Service:           contentService,
		EnrollmentService: enrollmentService,
		CourseService:     courseService,
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
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	var req CreateContentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	content, err := h.Service.Create(r.Context(), req.ToCommand(moduleID, user.ID))
	if err != nil {
		handleError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, content)
}

const maxUploadSize = 50 << 20 // 50 MB

func (h *ContentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}
	defer file.Close()

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		title = header.Filename
	}

	order := 0
	if orderStr := r.FormValue("order"); orderStr != "" {
		order, _ = strconv.Atoi(orderStr)
	}

	ext := filepath.Ext(header.Filename)
	storageName := fmt.Sprintf("%d/%d_%d%s", moduleID, time.Now().UnixNano(), user.ID, ext)

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	fileContent, err := h.Service.CreateFile(r.Context(), &services.CreateFileCommand{
		ModuleID: moduleID,
		Title:    title,
		Order:    order,
		FileName: storageName,
		FileSize: header.Size,
		MimeType: mimeType,
		UserID:   user.ID,
		Content:  file,
	})
	if err != nil {
		handleError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, fileContent)
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
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	_, err = strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	contentID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	var req UpdateContentRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.Service.Update(r.Context(), req.ToCommand(contentID, user.ID)); err != nil {
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	_, err = strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	contentID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
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
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) Publish(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	_, err = strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	contentID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
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
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
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
		handleError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	items, err := h.Service.List(r.Context(), &services.ListContentQuery{
		ModuleID: moduleID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		handleError(w, r, err)
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
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	_, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	contentID, err := strconv.ParseInt(chi.URLParam(r, "contentId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	content, err := h.Service.Get(r.Context(), &services.GetContentQuery{
		ContentID: contentID,
		ModuleID:  moduleID,
		UserID:    user.ID,
		UserRole:  user.Role,
	})
	if err != nil {
		handleError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, content)
}

func (h *ContentHandler) Download(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, r, errors.ErrInvalidCredentials)
		return
	}

	fileID, err := strconv.ParseInt(chi.URLParam(r, "fileId"), 10, 64)
	if err != nil {
		handleError(w, r, errors.ErrInvalidInput)
		return
	}

	file, err := h.Service.GetFileForDownload(r.Context(), &services.GetFileForDownloadQuery{
		FileID:          fileID,
		UserID:          user.ID,
		UserRole:        user.Role,
		EnrolledLearner: false,
	})
	if err == errors.ErrForbidden {
		file, err = h.Service.GetFileForDownload(r.Context(), &services.GetFileForDownloadQuery{
			FileID:          fileID,
			UserID:          user.ID,
			UserRole:        user.Role,
			EnrolledLearner: true,
		})
		if err != nil {
			handleError(w, r, err)
			return
		}
		module, _ := h.Service.Modules.GetByID(r.Context(), file.ModuleID)
		if module == nil {
			handleError(w, r, errors.ErrNotFound)
			return
		}
		course, err := h.CourseService.Get(r.Context(), &services.GetCourseQuery{
			CourseID: module.CourseID,
			UserID:   user.ID,
			UserRole: user.Role,
		})
		if err != nil {
			handleError(w, r, err)
			return
		}
		enrolled, err := h.EnrollmentService.IsEnrolled(r.Context(), &services.IsEnrolledQuery{
			CourseID: course.ID,
			UserID:   user.ID,
		})
		if err != nil || !enrolled {
			handleError(w, r, errors.ErrForbidden)
			return
		}
	}
	if err != nil {
		handleError(w, r, err)
		return
	}

	fileContent, err := h.Service.GetFileContent(r.Context(), file)
	if err != nil {
		handleError(w, r, err)
		return
	}
	defer fileContent.Close()

	w.Header().Set("Content-Type", file.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.FileName))
	w.Header().Set("Content-Length", strconv.FormatInt(file.FileSize, 10))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, fileContent)
}
