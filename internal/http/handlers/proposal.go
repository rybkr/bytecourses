package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
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

func (h *ProposalHandlers) Proposals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.postProposals(w, r)
	case http.MethodGet:
		h.getProposals(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ProposalHandlers) ProposalByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getProposalByID(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

type newProposalRequest struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

func (h *ProposalHandlers) postProposals(w http.ResponseWriter, r *http.Request) {
	var request newProposalRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	actor, ok := h.actor(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	p := domain.NewProposal(strings.TrimSpace(request.Title), strings.TrimSpace(request.Summary), actor.ID)
	if err := h.proposals.InsertProposal(r.Context(), p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ProposalHandlers) getProposals(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	out := h.proposals.GetProposalsByUserID(r.Context(), actor.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *ProposalHandlers) getProposalByID(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Path[len("/api/proposals/"):]
	if pidStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

    p, ok := h.proposals.GetProposalByID(r.Context(), pid)
    if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

    w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProposalHandlers) actor(r *http.Request) (*domain.User, bool) {
	c, err := r.Cookie("session")
	if err != nil {
		return nil, false
	}

	uid, ok := h.sessions.GetUserIDByToken(c.Value)
	if !ok {
		return nil, false
	}

	u, ok := h.users.GetUserByID(r.Context(), uid)
	return u, ok
}
