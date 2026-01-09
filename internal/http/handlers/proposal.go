package handlers

import (
	"bytecourses/internal/services"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type ProposalHandler struct {
	services *services.Services
}

func NewProposalHandler(services *services.Services) *ProposalHandler {
	return &ProposalHandler{
		services: services,
	}
}

type ActionRequest struct {
	ReviewNotes string `json:"review_notes"`
}

func (h *ProposalHandler) Create(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var request services.CreateProposalRequest
	if !decodeJSON(w, r, &request) {
		return
	}
	request.AuthorID = user.ID

	proposal, err := h.services.Proposals.CreateProposal(r.Context(), &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, proposal)
}

func (h *ProposalHandler) Get(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	p, ok := requireProposal(w, r)
	if !ok {
		return
	}

	proposal, err := h.services.Proposals.GetProposal(r.Context(), user, p)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, proposal)
}

func (h *ProposalHandler) List(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	proposals, err := h.services.Proposals.ListProposals(r.Context(), user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, proposals)
}

func (h *ProposalHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	proposals, err := h.services.Proposals.ListMyProposals(r.Context(), user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, proposals)
}

func (h *ProposalHandler) Update(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPatch) {
		return
	}

	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	p, ok := requireProposal(w, r)
	if !ok {
		return
	}

	var request services.UpdateProposalRequest
	if !decodeJSON(w, r, &request) {
		return
	}

	err := h.services.Proposals.UpdateProposal(r.Context(), p, user, &request)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) Delete(w http.ResponseWriter, r *http.Request) {
	u, ok := requireUser(w, r)
	if !ok {
		return
	}
	p, ok := requireProposal(w, r)
	if !ok {
		return
	}

	err := h.services.Proposals.DeleteProposal(r.Context(), p, u)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandler) Action(w http.ResponseWriter, r *http.Request) {
	action := chi.URLParam(r, "action")
	if action == "" {
		http.Error(w, "missing action", http.StatusBadRequest)
		return
	}

	u, ok := requireUser(w, r)
	if !ok {
		return
	}
	p, ok := requireProposal(w, r)
	if !ok {
		return
	}

	var request ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var err error
	switch action {
	case "submit":
		err = h.services.Proposals.SubmitProposal(r.Context(), p, u)

	case "withdraw":
		err = h.services.Proposals.WithdrawProposal(r.Context(), p, u)

	case "approve", "reject", "request-changes":
		err = h.services.Proposals.ReviewProposal(r.Context(), p, u, services.ReviewProposalRequest{
			Action: action,
			Notes:  request.ReviewNotes,
		})

	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
		return
	}

	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
