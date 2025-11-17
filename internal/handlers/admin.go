package handlers

import (
	"encoding/json"
	"github.com/rybkr/bytecourses/internal/models"
	"github.com/rybkr/bytecourses/internal/store"
	"log"
	"net/http"
	"strconv"
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) ApproveCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid course id for approval: %s", idStr)
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateCourseStatus(r.Context(), id, models.StatusApproved); err != nil {
		log.Printf("failed to approve course in admin handler: id=%d, error=%v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AdminHandler) RejectCourse(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid course id for rejection: %s", idStr)
		http.Error(w, "invalid course id", http.StatusBadRequest)
		return
	}

	if err := h.store.RejectCourse(r.Context(), id); err != nil {
		log.Printf("failed to reject course in admin handler: id=%d, error=%v", id, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
