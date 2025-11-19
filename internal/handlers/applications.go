package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/rybkr/bytecourses/internal/helpers"
	"github.com/rybkr/bytecourses/internal/middleware"
	"github.com/rybkr/bytecourses/internal/models"
	"github.com/rybkr/bytecourses/internal/store"
	"github.com/rybkr/bytecourses/internal/validation"
)

type ApplicationHandler struct {
	store *store.Store
}

func NewApplicationHandler(store *store.Store) *ApplicationHandler {
	return &ApplicationHandler{store: store}
}

func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		helpers.Unauthorized(w, "unauthorized")
		return
	}

	var app models.Application
	if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
		log.Printf("failed to decode application request: %v", err)
		helpers.BadRequest(w, "invalid request body")
		return
	}

	app.InstructorID = user.ID

	// Default to draft if status not provided
	if app.Status == "" {
		app.Status = models.StatusDraft
	}

	// Validate based on status
	v := validation.New()
	isDraft := app.Status == models.StatusDraft

	if isDraft {
		// For drafts: title can be empty, but if provided must be valid
		if app.Title != "" {
			v.MinLength(app.Title, 3, "title")
			v.MaxLength(app.Title, 255, "title")
		}
		v.Required(app.Description, "description")
	} else {
		// For submissions: full validation required
		v.Required(app.Title, "title")
		v.MinLength(app.Title, 3, "title")
		v.MaxLength(app.Title, 255, "title")
		v.Required(app.Description, "description")
		v.Required(app.LearningObjectives, "learning_objectives")
		v.Required(app.CourseFormat, "course_format")
		v.Required(app.SkillLevel, "skill_level")
	}

	// Check draft limit
	if isDraft {
		count, err := h.store.CountDraftsByInstructor(r.Context(), user.ID)
		if err != nil {
			log.Printf("failed to count drafts: %v", err)
			helpers.InternalServerError(w, "internal server error")
			return
		}
		if count >= 16 {
			helpers.BadRequest(w, "draft limit reached (maximum 16 drafts)")
			return
		}
	}

	if !v.Valid() {
		log.Printf("application validation failed: %v", v.Errors)
		helpers.ValidationError(w, v.Errors)
		return
	}

	if err := h.store.CreateApplication(r.Context(), &app); err != nil {
		log.Printf("failed to create application in handler: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.Created(w, app)
}

func (h *ApplicationHandler) GetMyApplications(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		helpers.Unauthorized(w, "unauthorized")
		return
	}

	applications, err := h.store.GetApplicationsByInstructor(r.Context(), user.ID)
	if err != nil {
		log.Printf("failed to get applications in handler: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.Success(w, applications)
}

func (h *ApplicationHandler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		helpers.Unauthorized(w, "unauthorized")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid application id: %s", idStr)
		helpers.BadRequest(w, "invalid application id")
		return
	}

	app, err := h.store.GetApplicationByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to get application: %v", err)
		helpers.NotFound(w, "application not found")
		return
	}

	if app.InstructorID != user.ID {
		log.Printf("user %d attempted to update application %d owned by %d", user.ID, id, app.InstructorID)
		helpers.Forbidden(w, "forbidden")
		return
	}

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		log.Printf("failed to decode update request: %v", err)
		helpers.BadRequest(w, "invalid request body")
		return
	}

	if err := h.store.UpdateApplication(r.Context(), id, updateData); err != nil {
		log.Printf("failed to update application: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.NoContent(w)
}

func (h *ApplicationHandler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		helpers.Unauthorized(w, "unauthorized")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid application id: %s", idStr)
		helpers.BadRequest(w, "invalid application id")
		return
	}

	app, err := h.store.GetApplicationByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to get application: %v", err)
		helpers.NotFound(w, "application not found")
		return
	}

	if app.InstructorID != user.ID {
		log.Printf("user %d attempted to delete application %d owned by %d", user.ID, id, app.InstructorID)
		helpers.Forbidden(w, "forbidden")
		return
	}

	if err := h.store.DeleteApplication(r.Context(), id); err != nil {
		log.Printf("failed to delete application: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.NoContent(w)
}

func (h *ApplicationHandler) SubmitApplication(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		log.Println("user not found in context")
		helpers.Unauthorized(w, "unauthorized")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("invalid application id: %s", idStr)
		helpers.BadRequest(w, "invalid application id")
		return
	}

	app, err := h.store.GetApplicationByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to get application: %v", err)
		helpers.NotFound(w, "application not found")
		return
	}

	if app.InstructorID != user.ID {
		log.Printf("user %d attempted to submit application %d owned by %d", user.ID, id, app.InstructorID)
		helpers.Forbidden(w, "forbidden")
		return
	}

	if app.Status != models.StatusDraft {
		helpers.BadRequest(w, "only draft applications can be submitted")
		return
	}

	// Decode request body to get updated form data
	var updateData map[string]interface{}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
			log.Printf("failed to decode submit request: %v", err)
			helpers.BadRequest(w, "invalid request body")
			return
		}
	}

	// Update application with provided data if any
	if len(updateData) > 0 {
		if err := h.store.UpdateApplication(r.Context(), id, updateData); err != nil {
			log.Printf("failed to update application before submission: %v", err)
			helpers.InternalServerError(w, "internal server error")
			return
		}

		// Re-fetch application to get updated values
		app, err = h.store.GetApplicationByID(r.Context(), id)
		if err != nil {
			log.Printf("failed to get updated application: %v", err)
			helpers.InternalServerError(w, "internal server error")
			return
		}
	}

	// Validate required fields for submission
	v := validation.New()
	v.Required(app.Title, "title")
	v.MinLength(app.Title, 3, "title")
	v.MaxLength(app.Title, 255, "title")
	v.Required(app.Description, "description")
	v.Required(app.LearningObjectives, "learning_objectives")
	v.Required(app.CourseFormat, "course_format")
	v.Required(app.SkillLevel, "skill_level")

	if !v.Valid() {
		log.Printf("application validation failed: %v", v.Errors)
		helpers.ValidationError(w, v.Errors)
		return
	}

	updates := map[string]interface{}{
		"status": models.StatusPending,
	}

	if err := h.store.UpdateApplication(r.Context(), id, updates); err != nil {
		log.Printf("failed to submit application: %v", err)
		helpers.InternalServerError(w, "internal server error")
		return
	}

	helpers.NoContent(w)
}

