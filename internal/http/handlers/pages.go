package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/store"
	"encoding/json"
	"net/http"
	"strconv"
)

type PageHandlers struct {
	users         store.UserStore
	sessions      auth.SessionStore
	proposals     store.ProposalStore
}

func NewPageHandlers(users store.UserStore, sessions auth.SessionStore, proposals store.ProposalStore) *PageHandlers {
	return &PageHandlers{
		users:     users,
		sessions:  sessions,
		proposals: proposals,
	}
}

func (h *PageHandlers) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := &TemplateData{Page: "home"}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) Login(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to home
	if _, ok := actorFromRequest(r, h.sessions, h.users); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := &TemplateData{Page: "login"}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) Register(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to home
	if _, ok := actorFromRequest(r, h.sessions, h.users); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := &TemplateData{Page: "register"}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) ProposalsList(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	user, ok := actorFromRequest(r, h.sessions, h.users)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	data := &TemplateData{User: &user, Page: "proposals"}
	Render(w, data)
}

func (h *PageHandlers) ProposalNew(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	user, ok := actorFromRequest(r, h.sessions, h.users)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	data := &TemplateData{User: &user, Page: "proposal_new"}
	Render(w, data)
}

func (h *PageHandlers) ProposalView(w http.ResponseWriter, r *http.Request) {
	// Require authentication
	user, ok := actorFromRequest(r, h.sessions, h.users)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	
	// Extract proposal ID from path
	pidStr := r.URL.Path[len("/proposals/"):]
	if pidStr == "" {
		http.NotFound(w, r)
		return
	}
	
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	
	// Get proposal
	p, ok := h.proposals.GetProposalByID(r.Context(), pid)
	if !ok || p.AuthorID != user.ID {
		http.NotFound(w, r)
		return
	}
	
	// Convert proposal to JSON for the template
	proposalJSON, _ := json.Marshal(p)
	
	data := &TemplateData{
		User:         &user,
		Proposal:     &p,
		ProposalJSON: string(proposalJSON),
		Page:         "proposal_view",
	}
	Render(w, data)
}

