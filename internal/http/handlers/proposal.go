package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
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

func (h *ProposalHandlers) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.Create(w, r)
	case http.MethodGet:
		h.ListMine(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ProposalHandlers) Item(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Get(w, r)
	case http.MethodPatch:
		h.Update(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ProposalHandlers) Action(w http.ResponseWriter, r *http.Request) {
    action := chi.URLParam(r, "action")
    switch action {
    case "submit":
        h.Submit(w, r)
    default:
        http.Error(w, "unknown action", http.StatusBadRequest)
    }
}

type ProposalCreateResponse struct {
	ID int64 `json:"id"`
}

func (h *ProposalHandlers) Create(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	user, ok := requireUser(w, r, h.sessions, h.users)
	if !ok {
		return
	}

	var p domain.Proposal
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p.AuthorID = user.ID
	// Set default status to draft if not provided
	if p.Status == "" {
		p.Status = domain.ProposalStatusDraft
	}
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

func (h *ProposalHandlers) ListMine(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	user, ok := requireUser(w, r, h.sessions, h.users)
	if !ok {
		return
	}

	response := h.proposals.GetProposalsByUserID(r.Context(), user.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProposalHandlers) Get(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}
	user, ok := requireUser(w, r, h.sessions, h.users)
	if !ok {
		return
	}

	pid, ok := h.requireProposalID(w, r)
	if !ok {
		return
	}

	p, ok := h.proposals.GetProposalByID(r.Context(), pid)
	if !ok || p.AuthorID != user.ID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProposalHandlers) Update(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPatch) {
		return
	}
	user, ok := requireUser(w, r, h.sessions, h.users)
	if !ok {
		return
	}
	pid, ok := h.requireProposalID(w, r)
	if !ok {
		return
	}

	var p domain.Proposal
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if p.AuthorID != user.ID {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	p.ID = pid

	if err := h.proposals.UpdateProposal(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *ProposalHandlers) Submit(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	user, ok := requireUser(w, r, h.sessions, h.users)
	if !ok {
		return
	}
	pid, ok := h.requireProposalID(w, r)
	if !ok {
		return
	}

	p, ok := h.proposals.GetProposalByID(r.Context(), pid)
	if !ok || p.AuthorID != user.ID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	p.ID = pid

	p.Status = domain.ProposalStatusSubmitted
	if err := h.proposals.UpdateProposal(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

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
