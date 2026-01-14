package handlers

import (
	"bytecourses/internal/services"
	"github.com/go-chi/chi/v5"
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

	proposal, err := h.services.Proposals.GetProposal(r.Context(), p, user)
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

	var err error
	switch action {
	case "submit":
		err = h.services.Proposals.SubmitProposal(r.Context(), p, u)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case "withdraw":
		err = h.services.Proposals.WithdrawProposal(r.Context(), p, u)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case "create-course":
		course, err := h.services.Courses.CreateCourseFromProposal(r.Context(), p, u)
		if err != nil {
			handleServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, course)
	case "approve", "reject", "request-changes":
		var request services.ProposalActionRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		err = h.services.Proposals.ReviewProposal(r.Context(), p, u, &services.ReviewProposalRequest{
			Action: action,
			Notes:  request.ReviewNotes,
		})
		if err != nil {
			handleServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "unknown action", http.StatusBadRequest)
		return
	}
}
