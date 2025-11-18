package handlers

import (
	"encoding/json"
	"github.com/rybkr/bytecourses/internal/middleware"
	"github.com/rybkr/bytecourses/internal/store"
	"log"
	"net/http"
	"strconv"
)

type UserHandler struct {
	store *store.Store
}

func NewUserHandler(store *store.Store) *UserHandler {
	return &UserHandler{store: store}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	fullUser, err := h.store.GetUserByID(r.Context(), user.ID)
	if err != nil {
		log.Printf("failed to get user profile: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fullUser)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid user id: %s", idStr)
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	user, err := h.store.GetUserByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to get user: %v", err)
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var updateData struct {
		Name string `json:"name"`
		Bio  string `json:"bio"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		log.Printf("failed to decode update request: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.store.UpdateUserProfile(r.Context(), user.ID, updateData.Name, updateData.Bio); err != nil {
		log.Printf("failed to update profile: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
