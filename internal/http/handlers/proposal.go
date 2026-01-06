package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/http/middleware"
	"bytecourses/internal/store"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
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

func (h *ProposalHandlers) WithUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.RequireUser(w, r, h.sessions, h.users)
		if !ok {
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *ProposalHandlers) WithAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.RequireAdminUser(w, r, h.sessions, h.users)
		if !ok {
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *ProposalHandlers) WithProposal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pid, ok := h.requireProposalID(w, r)
		if !ok {
			return
		}

		p, ok := h.proposals.GetProposalByID(r.Context(), pid)
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		// ensure ID if correct in case store doesn't populate it
		p.ID = pid

		ctx := context.WithValue(r.Context(), "proposal", p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type ActionRequest struct {
	ReviewNotes string `json:"review_notes"`
}

func (h *ProposalHandlers) Action(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r)
	p := proposalFrom(r)
	action := chi.URLParam(r, "action")

	var actionReq ActionRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&actionReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	switch action {
	case "submit":
		if p.AuthorID != user.ID {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if p.Status != domain.ProposalStatusDraft && p.Status != domain.ProposalStatusChangesRequested {
			http.Error(w, "invalid state", http.StatusConflict)
			return
		}
		p.Status = domain.ProposalStatusSubmitted

	case "withdraw":
		if p.AuthorID != user.ID {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if p.Status != domain.ProposalStatusSubmitted {
			http.Error(w, "invalid state", http.StatusConflict)
			return
		}
		p.Status = domain.ProposalStatusWithdrawn

	case "approve", "reject", "request-changes":
		if user.Role != domain.UserRoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if p.Status != domain.ProposalStatusSubmitted {
			http.Error(w, "invalid state", http.StatusConflict)
			return
		}
		p.ReviewerID = user.ID
		p.ReviewNotes = actionReq.ReviewNotes
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

	if err := h.proposals.UpdateProposal(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type ProposalCreateResponse struct {
	ID int64 `json:"id"`
}

func (h *ProposalHandlers) Create(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r)

	var p domain.Proposal
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.AuthorID = user.ID
	p.Status = domain.ProposalStatusDraft

	if err := h.proposals.InsertProposal(r.Context(), &p); err != nil {
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
	user := userFrom(r)

	switch user.Role {
	// If the user is an admin, then GET /api/proposals shall return all proposals submitted for review.
	case domain.UserRoleAdmin:
		response := h.proposals.GetAllSubmittedProposals(r.Context())
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	// Else, it shall return all proposals owned by the user.
	default:
		http.Redirect(w, r, "/api/proposals/mine", http.StatusSeeOther)
	}
}

func (h *ProposalHandlers) ListMine(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r)

	response := h.proposals.GetProposalsByUserID(r.Context(), user.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProposalHandlers) Get(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r)

	pid, ok := h.requireProposalID(w, r)
	if !ok {
		return
	}

	p, ok := h.proposals.GetProposalByID(r.Context(), pid)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if user.Role != domain.UserRoleAdmin && p.AuthorID != user.ID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if user.Role == domain.UserRoleAdmin &&
		p.Status != domain.ProposalStatusSubmitted &&
		p.Status != domain.ProposalStatusApproved &&
		p.Status != domain.ProposalStatusRejected &&
		p.Status != domain.ProposalStatusChangesRequested {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProposalHandlers) Update(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r)
	p := proposalFrom(r)

	if p.AuthorID != user.ID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if p.Status != domain.ProposalStatusDraft && p.Status != domain.ProposalStatusChangesRequested {
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

	if err := h.proposals.UpdateProposal(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandlers) Delete(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r)
	p := proposalFrom(r)

	if p.AuthorID != user.ID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if err := h.proposals.DeleteProposal(r.Context(), p.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProposalHandlers) Approve(w http.ResponseWriter, r *http.Request) {
	user := userFrom(r)
	pid, ok := h.requireProposalID(w, r)
	if !ok {
		return
	}

	p, ok := h.proposals.GetProposalByID(r.Context(), pid)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
	}
	p.ID = pid

	p.Status = domain.ProposalStatusApproved
	p.ReviewerID = user.ID

}

func userFrom(r *http.Request) domain.User {
	return r.Context().Value("user").(domain.User)
}

func proposalFrom(r *http.Request) domain.Proposal {
	return r.Context().Value("proposal").(domain.Proposal)
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
