package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
    "github.com/rybkr/bytecourses/internal/helpers"
    "github.com/rybkr/bytecourses/internal/middleware"
    "github.com/rybkr/bytecourses/internal/store"
    "github.com/rybkr/bytecourses/internal/validation"
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
        helpers.Error(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    
    fullUser, err := h.store.GetUserByID(r.Context(), user.ID)
    if err != nil {
        log.Printf("failed to get user profile: %v", err)
        helpers.Error(w, http.StatusInternalServerError, "internal server error")
        return
    }
    
    helpers.Success(w, fullUser)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
    idStr := r.URL.Query().Get("id")
    id, err := strconv.Atoi(idStr)
    if err != nil {
        log.Printf("invalid user id: %s", idStr)
        helpers.Error(w, http.StatusBadRequest, "invalid user id")
        return
    }
    
    user, err := h.store.GetUserByID(r.Context(), id)
    if err != nil {
        log.Printf("failed to get user: %v", err)
        helpers.Error(w, http.StatusNotFound, "user not found")
        return
    }
    
    helpers.Success(w, user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
    user, ok := middleware.GetUserFromContext(r.Context())
    if !ok {
        log.Println("user not found in context")
        helpers.Error(w, http.StatusUnauthorized, "unauthorized")
        return
    }
    
    var updateData struct {
        Name string `json:"name"`
        Bio  string `json:"bio"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
        log.Printf("failed to decode update request: %v", err)
        helpers.Error(w, http.StatusBadRequest, "invalid request body")
        return
    }
    
    v := validation.New()
    v.MaxLength(updateData.Name, 255, "name")
    v.MaxLength(updateData.Bio, 1000, "bio")
    
    if !v.Valid() {
        helpers.JSON(w, http.StatusBadRequest, map[string]interface{}{
            "error": "validation failed",
            "fields": v.Errors,
        })
        return
    }
    
    if err := h.store.UpdateUserProfile(r.Context(), user.ID, updateData.Name, updateData.Bio); err != nil {
        log.Printf("failed to update profile: %v", err)
        helpers.Error(w, http.StatusInternalServerError, "internal server error")
        return
    }
    
    helpers.NoContent(w)
}
