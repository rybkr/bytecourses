package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"encoding/json"
	"net/http"
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

    p := domain.NewProposal(strings.TrimSpace(request.Title), strings.TrimSpace(request.Summary))
    if err := h.proposals.InsertProposal(r.Context(), p); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusOK)
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
