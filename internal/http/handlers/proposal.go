package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/http/middleware"
	"bytecourses/internal/store"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
)

type ProposalHandlers struct {
	proposals store.ProposalStore
	users     store.UserStore
	sessions  auth.SessionStore
}

func NewProposalHandlers(proposals store.ProposalStore, users store.UserStore, sessions auth.SessionStore) *ProposalHandlers {
	return &ProposalHandlers{
		proposals: proposals,
		users:     users,
		sessions:  sessions,
	}
}

type ActionRequest struct {
	ReviewNotes string `json:"review_notes"`
}

func (h *ProposalHandlers) Action(w http.ResponseWriter, r *http.Request) {
	action := chi.URLParam(r, "action")
	if action == "" {
		http.Error(w, "missing action", http.StatusBadRequest)
		return
	}

	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	p, ok := middleware.ProposalFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var request ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil && err != io.EOF {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	switch action {
	case "submit":
		if !p.IsOwnedBy(u) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if !p.IsAmendable() {
			http.Error(w, "invalid state", http.StatusConflict)
			return
		}
		p.Status = domain.ProposalStatusSubmitted

	case "withdraw":
		if !p.IsOwnedBy(u) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if p.Status != domain.ProposalStatusSubmitted {
			http.Error(w, "invalid state", http.StatusConflict)
			return
		}
		p.Status = domain.ProposalStatusWithdrawn

	case "approve", "reject", "request-changes":
		if !u.IsAdmin() {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if p.Status != domain.ProposalStatusSubmitted {
			http.Error(w, "invalid state", http.StatusConflict)
			return
		}
		p.ReviewerID = &u.ID
		p.ReviewNotes = request.ReviewNotes
		if action == "approve" {
			p.Status = domain.ProposalStatusApproved
		} else if action == "reject" {
			p.Status = domain.ProposalStatusRejected
		} else {
			p.Status = domain.ProposalStatusChangesRequested
		}

	default:
		http.Error(w, "unknown action", http.StatusNotFound)
		return
	}

	if err := h.proposals.UpdateProposal(r.Context(), p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type ProposalCreateResponse struct {
	ID int64 `json:"id"`
}

func (h *ProposalHandlers) Create(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var p domain.Proposal
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.AuthorID = u.ID
	p.Status = domain.ProposalStatusDraft

	if err := h.proposals.CreateProposal(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := ProposalCreateResponse{
		ID: p.ID,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *ProposalHandlers) List(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	switch u.Role {
	// If the user is an admin, then GET /api/proposals shall return all proposals submitted for review.
	case domain.UserRoleAdmin:
		response, _ := h.proposals.ListAllSubmittedProposals(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	// Else, it shall return all proposals owned by the user.
	default:
		response, _ := h.proposals.ListProposalsByAuthorID(r.Context(), u.ID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func (h *ProposalHandlers) ListMine(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	response, _ := h.proposals.ListProposalsByAuthorID(r.Context(), u.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProposalHandlers) Get(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	p, ok := middleware.ProposalFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !p.IsViewableBy(u) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProposalHandlers) Update(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	p, ok := middleware.ProposalFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !p.IsOwnedBy(u) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if !p.IsAmendable() {
		http.Error(w, "invalid state", http.StatusConflict)
		return
	}

	var patch domain.Proposal
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.Title = patch.Title
	p.Summary = patch.Summary
	p.TargetAudience = patch.TargetAudience
	p.LearningObjectives = patch.LearningObjectives
	p.Outline = patch.Outline
	p.AssumedPrerequisites = patch.AssumedPrerequisites

	if err := h.proposals.UpdateProposal(r.Context(), p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	p, ok := middleware.ProposalFromContext(r.Context())
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if !p.IsOwnedBy(u) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := h.proposals.DeleteProposalByID(r.Context(), p.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandlers) requireProposalID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	pidStr := chi.URLParam(r, "id")
	if pidStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return 0, false
	}

	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return 0, false
	}

	return pid, true
}
