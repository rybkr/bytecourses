package handlers

import (
	"bytecourses/internal/services"
	"net/http"
)

type ContentHandler struct {
	services *services.Services
}

func NewContentHandler(services *services.Services) *ContentHandler {
	return &ContentHandler{
		services: services,
	}
}

func (h *ContentHandler) CreateLecture(w http.ResponseWriter, r *http.Request) {
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
	module, ok := requireModule(w, r)
	if !ok {
		return
	}

	var request services.CreateLectureRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	item, err := h.services.Content.CreateLecture(r.Context(), module, course, user, &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *ContentHandler) GetLecture(w http.ResponseWriter, r *http.Request) {
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
	item, ok := requireContentItem(w, r)
	if !ok {
		return
	}

	contentItem, lecture, err := h.services.Content.GetLecture(r.Context(), item, module, course, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"item":    contentItem,
		"lecture": lecture,
	})
}

func (h *ContentHandler) UpdateLecture(w http.ResponseWriter, r *http.Request) {
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
	item, ok := requireContentItem(w, r)
	if !ok {
		return
	}

	var request services.UpdateLectureRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	err := h.services.Content.UpdateLecture(r.Context(), item, module, course, user, &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) DeleteContent(w http.ResponseWriter, r *http.Request) {
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
	item, ok := requireContentItem(w, r)
	if !ok {
		return
	}

	err := h.services.Content.DeleteContent(r.Context(), item, module, course, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) ListContent(w http.ResponseWriter, r *http.Request) {
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

	items, lectures, err := h.services.Content.ListContent(r.Context(), module, course, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":    items,
		"lectures": lectures,
	})
}

func (h *ContentHandler) ReorderContent(w http.ResponseWriter, r *http.Request) {
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
	module, ok := requireModule(w, r)
	if !ok {
		return
	}

	var request services.ReorderContentRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	err := h.services.Content.ReorderContent(r.Context(), module, course, user, &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) PublishContent(w http.ResponseWriter, r *http.Request) {
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
	module, ok := requireModule(w, r)
	if !ok {
		return
	}
	item, ok := requireContentItem(w, r)
	if !ok {
		return
	}

	err := h.services.Content.PublishContent(r.Context(), item, module, course, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ContentHandler) UnpublishContent(w http.ResponseWriter, r *http.Request) {
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
	module, ok := requireModule(w, r)
	if !ok {
		return
	}
	item, ok := requireContentItem(w, r)
	if !ok {
		return
	}

	err := h.services.Content.UnpublishContent(r.Context(), item, module, course, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
