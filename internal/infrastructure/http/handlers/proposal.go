package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"bytecourses/internal/domain"
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
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	var request CreateProposalRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	p, err := h.Service.Create(r.Context(), request.ToCommand(user.ID))
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, p)
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
		UserID:               userID,
		Title:                strings.TrimSpace(r.Title),
		Summary:              strings.TrimSpace(r.Summary),
		Qualifications:       strings.TrimSpace(r.Qualifications),
		TargetAudience:       strings.TrimSpace(r.TargetAudience),
		LearningObjectives:   strings.TrimSpace(r.LearningObjectives),
		Outline:              strings.TrimSpace(r.Outline),
		AssumedPrerequisites: strings.TrimSpace(r.AssumedPrerequisites),
	}
}

func (h *ProposalHandler) Update(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var request UpdateProposalRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	_, err = h.Service.GetByID(r.Context(), &services.GetProposalByIDQuery{
		ProposalID: id,
		UserID:     user.ID,
		IsAdmin:    user.IsAdmin(),
	})
	if err != nil {
		handleError(w, err)
		return
	}

	_, err = h.Service.Update(r.Context(), request.ToCommand(id, user.ID))
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) Submit(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	_, err = h.Service.Submit(r.Context(), &services.SubmitProposalCommand{
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
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	_, err = h.Service.Withdraw(r.Context(), &services.WithdrawProposalCommand{
		ProposalID: id,
		UserID:     user.ID,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type ReviewProposalRequest struct {
	Decision string `json:"decision"`
	Notes    string `json:"notes"`
}

func (h *ProposalHandler) Review(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var request ReviewProposalRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	p, err := h.Service.Review(r.Context(), &services.ReviewProposalCommand{
		ProposalID: id,
		ReviewerID: user.ID,
		Decision:   services.ReviewDecision(request.Decision),
		Notes:      request.Notes,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *ProposalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.Service.Delete(r.Context(), &services.DeleteProposalCommand{
		ProposalID: id,
		UserID:     user.ID,
	}); err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, err := h.Service.GetByID(r.Context(), &services.GetProposalByIDQuery{
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
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	proposals, err := h.Service.ListAll(r.Context(), user.IsAdmin())
	if err != nil {
		handleError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, proposals)
}

func (h *ProposalHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	proposals, err := h.Service.ListMine(r.Context(), user.ID)
	if err != nil {
		handleError(w, err)
		return
	}

	if proposals == nil {
		proposals = []domain.Proposal{}
	}
	writeJSON(w, http.StatusOK, proposals)
}

type ReviewActionRequest struct {
	Notes string `json:"notes"`
}

func (h *ProposalHandler) Approve(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var request ReviewActionRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	_, err = h.Service.Review(r.Context(), &services.ReviewProposalCommand{
		ProposalID: id,
		ReviewerID: user.ID,
		Decision:   services.ReviewDecisionApprove,
		Notes:      request.Notes,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) Reject(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var request ReviewActionRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	_, err = h.Service.Review(r.Context(), &services.ReviewProposalCommand{
		ProposalID: id,
		ReviewerID: user.ID,
		Decision:   services.ReviewDecisionReject,
		Notes:      request.Notes,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) RequestChanges(w http.ResponseWriter, r *http.Request) {
	user, ok := requireAuthenticatedUser(w, r)
	if !ok {
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var request ReviewActionRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	_, err = h.Service.Review(r.Context(), &services.ReviewProposalCommand{
		ProposalID: id,
		ReviewerID: user.ID,
		Decision:   services.ReviewDecisionRequestChanges,
		Notes:      request.Notes,
	})
	if err != nil {
		handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
