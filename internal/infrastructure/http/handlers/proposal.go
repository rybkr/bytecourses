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

type ProposalHandler struct {
	Service *services.ProposalService
}

func NewProposalHandler(proposalService *services.ProposalService) *ProposalHandler {
	return &ProposalHandler{
		Service: proposalService,
	}
}

type CreateProposalRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	Qualifications       string `json:"qualifications"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	Outline              string `json:"outline"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (r *CreateProposalRequest) ToCommand(authorID int64) *services.CreateProposalCommand {
	return &services.CreateProposalCommand{
		AuthorID:             authorID,
		Title:                strings.TrimSpace(r.Title),
		Summary:              strings.TrimSpace(r.Summary),
		Qualifications:       strings.TrimSpace(r.Qualifications),
		TargetAudience:       strings.TrimSpace(r.TargetAudience),
		LearningObjectives:   strings.TrimSpace(r.LearningObjectives),
		Outline:              strings.TrimSpace(r.Outline),
		AssumedPrerequisites: strings.TrimSpace(r.AssumedPrerequisites),
	}
}

func (h *ProposalHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	var req CreateProposalRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	proposal, err := h.Service.Create(r.Context(), req.ToCommand(user.ID))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, proposal)
}

type UpdateProposalRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	Qualifications       string `json:"qualifications"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	Outline              string `json:"outline"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (r *UpdateProposalRequest) ToCommand(proposalID, userID int64) *services.UpdateProposalCommand {
	return &services.UpdateProposalCommand{
		ProposalID:           proposalID,
		Title:                strings.TrimSpace(r.Title),
		Summary:              strings.TrimSpace(r.Summary),
		Qualifications:       strings.TrimSpace(r.Qualifications),
		TargetAudience:       strings.TrimSpace(r.TargetAudience),
		LearningObjectives:   strings.TrimSpace(r.LearningObjectives),
		Outline:              strings.TrimSpace(r.Outline),
		AssumedPrerequisites: strings.TrimSpace(r.AssumedPrerequisites),
		UserID:               userID,
	}
}

func (h *ProposalHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	proposalID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req UpdateProposalRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.Service.Update(r.Context(), req.ToCommand(proposalID, user.ID)); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type UpdateProposalStatusRequest struct {
	Status      domain.ProposalStatus `json:"status"`
	ReviewNotes string                `json:"review_notes"`
}

func (r *UpdateProposalStatusRequest) ToCommand(proposalID, userID int64, userRole domain.UserRole) *services.UpdateProposalStatusCommand {
	return &services.UpdateProposalStatusCommand{
		ProposalID:  proposalID,
		Status:      r.Status,
		ReviewNotes: strings.TrimSpace(r.ReviewNotes),
		UserID:      userID,
		UserRole:    userRole,
	}
}

func (h *ProposalHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	proposalID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req UpdateProposalStatusRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if err := h.Service.UpdateStatus(r.Context(), req.ToCommand(proposalID, user.ID, user.Role)); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	proposalID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), &services.DeleteProposalCommand{
		ProposalID: proposalID,
		UserID:     user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) Get(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	proposalID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	proposal, err := h.Service.Get(r.Context(), &services.GetProposalQuery{
		ProposalID: proposalID,
		UserID:     user.ID,
		UserRole:   user.Role,
	})
    if err != nil {
        handleError(w, err)
        return
    }

    writeJSON(w, http.StatusOK, proposal)
}

func (h *ProposalHandler) List(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		handleError(w, errors.ErrInvalidCredentials)
		return
	}

	proposals, err := h.Service.List(r.Context(), &services.ListProposalsQuery{
		UserID:   user.ID,
		UserRole: user.Role,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, proposals)
}
