package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/rybkr/bytecourses/internal/helpers"
	"github.com/rybkr/bytecourses/internal/store"
)

type AdminHandler struct {
	store *store.Store
}

func NewAdminHandler(store *store.Store) *AdminHandler {
	return &AdminHandler{store: store}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.store.GetAllUsers(r.Context())
	if err != nil {
		log.Printf("failed to get users in admin handler: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.Success(w, users)
}

func (h *AdminHandler) ListPendingApplications(w http.ResponseWriter, r *http.Request) {
	applications, err := h.store.GetPendingApplications(r.Context())
	if err != nil {
		log.Printf("failed to get pending applications in admin handler: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.Success(w, applications)
}

func (h *AdminHandler) ApproveApplication(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid application id for approval: %s", idStr)
		helpers.BadRequest(w, "invalid application id")
		return
	}

	app, err := h.store.ApproveApplication(r.Context(), id)
	if err != nil {
		log.Printf("failed to approve application in admin handler: id=%d, error=%v", id, err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	// Create course from application
	course, err := h.store.CreateCourseFromApplication(r.Context(), app)
	if err != nil {
		log.Printf("failed to create course from application: id=%d, error=%v", id, err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	// Delete application after creating course
	if err := h.store.DeleteApplication(r.Context(), id); err != nil {
		log.Printf("failed to delete application after approval: id=%d, error=%v", id, err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.Success(w, course)
}

func (h *AdminHandler) RejectApplication(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid application id for rejection: %s", idStr)
		helpers.BadRequest(w, "invalid application id")
		return
	}

	if err := h.store.RejectApplication(r.Context(), id); err != nil {
		log.Printf("failed to reject application in admin handler: id=%d, error=%v", id, err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.NoContent(w)
}
