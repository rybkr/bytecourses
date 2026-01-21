package handlers

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CourseHandler struct {
	Service *services.CourseService
}

func NewCourseHandler(courseService *services.CourseService) *CourseHandler {
	return &CourseHandler{
		Service: courseService,
	}
}

type CreateCourseRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	var request CreateCourseRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	c, err := h.Service.Create(r.Context(), &services.CreateCourseCommand{
		InstructorID:         user.ID,
		Title:                request.Title,
		Summary:              request.Summary,
		TargetAudience:       request.TargetAudience,
		LearningObjectives:   request.LearningObjectives,
		AssumedPrerequisites: request.AssumedPrerequisites,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

type UpdateCourseRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (h *CourseHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var request UpdateCourseRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	_, err = h.Service.Update(r.Context(), &services.UpdateCourseCommand{
		CourseID:             id,
		UserID:               user.ID,
		Title:                request.Title,
		Summary:              request.Summary,
		TargetAudience:       request.TargetAudience,
		LearningObjectives:   request.LearningObjectives,
		AssumedPrerequisites: request.AssumedPrerequisites,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CourseHandler) Publish(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	c, err := h.Service.Publish(r.Context(), &services.PublishCourseCommand{
		CourseID: id,
		UserID:   user.ID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, c)
}

type CreateFromProposalRequest struct {
	ProposalID int64 `json:"proposal_id"`
}

func (h *CourseHandler) CreateFromProposal(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	var request CreateFromProposalRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	c, err := h.Service.CreateFromProposal(r.Context(), &services.CreateCourseFromProposalCommand{
		ProposalID: request.ProposalID,
		UserID:     user.ID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

func (h *CourseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	c, err := h.Service.GetByID(r.Context(), &services.GetCourseByIDQuery{
		CourseID: id,
		UserID:   user.ID,
		IsAdmin:  user.IsAdmin(),
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, c)
}

func (h *CourseHandler) ListLive(w http.ResponseWriter, r *http.Request) {
	courses, err := h.Service.ListLive(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	if courses == nil {
		courses = []domain.Course{}
	}

	writeJSON(w, http.StatusOK, courses)
}
