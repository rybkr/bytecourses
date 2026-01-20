package handlers

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CourseHandler struct {
	courseService *services.CourseService
}

func NewCourseHandler(courseService *services.CourseService) *CourseHandler {
	return &CourseHandler{
		courseService: courseService,
	}
}

type createCourseRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var req createCourseRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	c, err := h.courseService.Create(r.Context(), &services.CreateCourseInput{
		InstructorID:         user.ID,
		Title:                req.Title,
		Summary:              req.Summary,
		TargetAudience:       req.TargetAudience,
		LearningObjectives:   req.LearningObjectives,
		AssumedPrerequisites: req.AssumedPrerequisites,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

type updateCourseRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (h *CourseHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updateCourseRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	_, err = h.courseService.Update(r.Context(), &services.UpdateCourseInput{
		CourseID:             id,
		UserID:               user.ID,
		Title:                req.Title,
		Summary:              req.Summary,
		TargetAudience:       req.TargetAudience,
		LearningObjectives:   req.LearningObjectives,
		AssumedPrerequisites: req.AssumedPrerequisites,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *CourseHandler) Publish(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	c, err := h.courseService.Publish(r.Context(), &services.PublishCourseInput{
		CourseID: id,
		UserID:   user.ID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, c)
}

type createFromProposalRequest struct {
	ProposalID int64 `json:"proposal_id"`
}

func (h *CourseHandler) CreateFromProposal(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var req createFromProposalRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	c, err := h.courseService.CreateFromProposal(r.Context(), &services.CreateFromProposalInput{
		ProposalID: req.ProposalID,
		UserID:     user.ID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, c)
}

func (h *CourseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	c, err := h.courseService.GetByID(r.Context(), &services.GetCourseByIDInput{
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
	courses, err := h.courseService.ListLive(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}

	// Ensure we return [] instead of null for empty list
	if courses == nil {
		courses = []domain.Course{}
	}
	writeJSON(w, http.StatusOK, courses)
}
