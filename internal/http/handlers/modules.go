package handlers

import (
	"bytecourses/internal/services"
	"net/http"
)

type ModuleHandler struct {
	services *services.Services
}

func NewModuleHandler(services *services.Services) *ModuleHandler {
	return &ModuleHandler{
		services: services,
	}
}

func (h *ModuleHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	course, ok := requireCourse(w, r)
	if !ok {
		return
	}

	var request services.CreateModuleRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	module, err := h.services.Modules.CreateModule(r.Context(), course, user, &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, module)
}

func (h *ModuleHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	course, ok := requireCourse(w, r)
	if !ok {
		return
	}
	module, ok := requireModule(w, r)
	if !ok {
		return
	}

	m, err := h.services.Modules.GetModule(r.Context(), module, course, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, m)
}

func (h *ModuleHandler) List(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	course, ok := requireCourse(w, r)
	if !ok {
		return
	}

	modules, err := h.services.Modules.ListModules(r.Context(), course, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, modules)
}

func (h *ModuleHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPatch) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	course, ok := requireCourse(w, r)
	if !ok {
		return
	}
	module, ok := requireModule(w, r)
	if !ok {
		return
	}

	var request services.UpdateModuleRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	err := h.services.Modules.UpdateModule(r.Context(), module, course, user, &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ModuleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodDelete) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	course, ok := requireCourse(w, r)
	if !ok {
		return
	}
	module, ok := requireModule(w, r)
	if !ok {
		return
	}

	err := h.services.Modules.DeleteModule(r.Context(), module, course, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ModuleHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	course, ok := requireCourse(w, r)
	if !ok {
		return
	}

	var request services.ReorderModulesRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	err := h.services.Modules.ReorderModules(r.Context(), course, user, &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
