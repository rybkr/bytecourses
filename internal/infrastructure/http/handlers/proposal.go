package handlers

import (
	"bytecourses/internal/domain"
	"bytecourses/internal/services"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProposalHandler struct {
	proposalService *services.ProposalService
}

func NewProposalHandler(proposalService *services.ProposalService) *ProposalHandler {
	return &ProposalHandler{
		proposalService: proposalService,
	}
}

type createProposalRequest struct {
	Title                string `json:"title"`
	Summary              string `json:"summary"`
	Qualifications       string `json:"qualifications"`
	TargetAudience       string `json:"target_audience"`
	LearningObjectives   string `json:"learning_objectives"`
	Outline              string `json:"outline"`
	AssumedPrerequisites string `json:"assumed_prerequisites"`
}

func (h *ProposalHandler) Create(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var req createProposalRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	p, err := h.proposalService.Create(r.Context(), &services.CreateProposalInput{
		AuthorID:             user.ID,
		Title:                req.Title,
		Summary:              req.Summary,
		Qualifications:       req.Qualifications,
		TargetAudience:       req.TargetAudience,
		LearningObjectives:   req.LearningObjectives,
		Outline:              req.Outline,
		AssumedPrerequisites: req.AssumedPrerequisites,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

type updateProposalRequest struct {
	Title                *string `json:"title"`
	Summary              *string `json:"summary"`
	Qualifications       *string `json:"qualifications"`
	TargetAudience       *string `json:"target_audience"`
	LearningObjectives   *string `json:"learning_objectives"`
	Outline              *string `json:"outline"`
	AssumedPrerequisites *string `json:"assumed_prerequisites"`
}

func (h *ProposalHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updateProposalRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	// First get the existing proposal to support partial updates
	existing, err := h.proposalService.GetByID(r.Context(), &services.GetByIDInput{
		ProposalID: id,
		UserID:     user.ID,
		IsAdmin:    user.IsAdmin(),
	})
	if err != nil {
		handleError(w, err)
		return
	}

	// Merge request fields with existing proposal
	input := &services.UpdateProposalInput{
		ProposalID:           id,
		UserID:               user.ID,
		Title:                existing.Title,
		Summary:              existing.Summary,
		Qualifications:       existing.Qualifications,
		TargetAudience:       existing.TargetAudience,
		LearningObjectives:   existing.LearningObjectives,
		Outline:              existing.Outline,
		AssumedPrerequisites: existing.AssumedPrerequisites,
	}

	// Override with provided values
	if req.Title != nil {
		input.Title = *req.Title
	}
	if req.Summary != nil {
		input.Summary = *req.Summary
	}
	if req.Qualifications != nil {
		input.Qualifications = *req.Qualifications
	}
	if req.TargetAudience != nil {
		input.TargetAudience = *req.TargetAudience
	}
	if req.LearningObjectives != nil {
		input.LearningObjectives = *req.LearningObjectives
	}
	if req.Outline != nil {
		input.Outline = *req.Outline
	}
	if req.AssumedPrerequisites != nil {
		input.AssumedPrerequisites = *req.AssumedPrerequisites
	}

	_, err = h.proposalService.Update(r.Context(), input)
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) Submit(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	_, err = h.proposalService.Submit(r.Context(), &services.SubmitProposalInput{
		ProposalID: id,
		UserID:     user.ID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	_, err = h.proposalService.Withdraw(r.Context(), &services.WithdrawProposalInput{
		ProposalID: id,
		UserID:     user.ID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type reviewProposalRequest struct {
	Decision string `json:"decision"`
	Notes    string `json:"notes"`
}

func (h *ProposalHandler) Review(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req reviewProposalRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	p, err := h.proposalService.Review(r.Context(), &services.ReviewProposalInput{
		ProposalID: id,
		ReviewerID: user.ID,
		Decision:   services.ReviewDecision(req.Decision),
		Notes:      req.Notes,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *ProposalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.proposalService.Delete(r.Context(), &services.DeleteProposalInput{
		ProposalID: id,
		UserID:     user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := h.proposalService.GetByID(r.Context(), &services.GetByIDInput{
		ProposalID: id,
		UserID:     user.ID,
		IsAdmin:    user.IsAdmin(),
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *ProposalHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	proposals, err := h.proposalService.ListAll(r.Context(), user.IsAdmin())
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, proposals)
}

func (h *ProposalHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	proposals, err := h.proposalService.ListMine(r.Context(), user.ID)
	if err != nil {
		handleError(w, err)
		return
	}

	// Ensure we return [] instead of null for empty list
	if proposals == nil {
		proposals = []domain.Proposal{}
	}
	writeJSON(w, http.StatusOK, proposals)
}

type reviewActionRequest struct {
	Notes string `json:"notes"`
}

func (h *ProposalHandler) Approve(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req reviewActionRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	_, err = h.proposalService.Review(r.Context(), &services.ReviewProposalInput{
		ProposalID: id,
		ReviewerID: user.ID,
		Decision:   services.ReviewDecisionApprove,
		Notes:      req.Notes,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) Reject(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req reviewActionRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	_, err = h.proposalService.Review(r.Context(), &services.ReviewProposalInput{
		ProposalID: id,
		ReviewerID: user.ID,
		Decision:   services.ReviewDecisionReject,
		Notes:      req.Notes,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) RequestChanges(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req reviewActionRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	_, err = h.proposalService.Review(r.Context(), &services.ReviewProposalInput{
		ProposalID: id,
		ReviewerID: user.ID,
		Decision:   services.ReviewDecisionRequestChanges,
		Notes:      req.Notes,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
