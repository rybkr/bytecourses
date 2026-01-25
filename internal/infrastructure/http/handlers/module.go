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

type ModuleHandler struct {
	Service *services.ModuleService
}

func NewModuleHandler(moduleService *services.ModuleService) *ModuleHandler {
	return &ModuleHandler{
		Service: moduleService,
	}
}

type CreateModuleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

func (r *CreateModuleRequest) ToCommand(courseID, userID int64) *services.CreateModuleCommand {
	return &services.CreateModuleCommand{
		CourseID:    courseID,
		Title:       strings.TrimSpace(r.Title),
		Description: strings.TrimSpace(r.Description),
		Order:       r.Order,
		UserID:      userID,
	}
}

func (h *ModuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	var req CreateModuleRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	module, err := h.Service.Create(r.Context(), req.ToCommand(courseID, user.ID))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, module)
}

type UpdateModuleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

func (r *UpdateModuleRequest) ToCommand(moduleID, userID int64) *services.UpdateModuleCommand {
	return &services.UpdateModuleCommand{
		ModuleID:    moduleID,
		Title:       strings.TrimSpace(r.Title),
		Description: strings.TrimSpace(r.Description),
		Order:       r.Order,
		UserID:      userID,
	}
}

func (h *ModuleHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	var req UpdateModuleRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.Service.Update(r.Context(), req.ToCommand(moduleID, user.ID)); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ModuleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), &services.DeleteModuleCommand{
		ModuleID: moduleID,
		UserID:   user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ModuleHandler) Publish(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Publish(r.Context(), &services.PublishModuleCommand{
		ModuleID: moduleID,
		UserID:   user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ModuleHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	modules, err := h.Service.List(r.Context(), &services.ListModulesQuery{
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		handleError(w, err)
		return
	}
	if modules == nil {
		modules = make([]domain.Module, 0)
	}

	writeJSON(w, http.StatusOK, modules)
}

func (h *ModuleHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	moduleID, err := strconv.ParseInt(chi.URLParam(r, "moduleId"), 10, 64)
	if err != nil {
		http.Error(w, "invalid module id", http.StatusBadRequest)
		return
	}

	module, err := h.Service.Get(r.Context(), &services.GetModuleQuery{
		ModuleID: moduleID,
		CourseID: courseID,
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, module)
}
