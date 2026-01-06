package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/http/middleware"
	"bytecourses/internal/store"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type PageHandlers struct {
	users     store.UserStore
	sessions  auth.SessionStore
	proposals store.ProposalStore
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
	data := &TemplateData{Page: "home.html"}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if _, ok := actorFromRequest(r, h.sessions, h.users); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := &TemplateData{Page: "login.html"}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if _, ok := actorFromRequest(r, h.sessions, h.users); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := &TemplateData{Page: "register.html"}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) ProposalsList(w http.ResponseWriter, r *http.Request) {
	user, ok := actorFromRequest(r, h.sessions, h.users)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if user.Role != domain.UserRoleAdmin {
		http.Redirect(w, r, "/proposals/mine", http.StatusSeeOther)
		return
	}

	data := &TemplateData{User: &user, Page: "proposals.html"}
	Render(w, data)
}

func (h *PageHandlers) ProposalsListMine(w http.ResponseWriter, r *http.Request) {
	user, ok := actorFromRequest(r, h.sessions, h.users)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := &TemplateData{User: &user, Page: "proposals.html"}
	Render(w, data)
}

func (h *PageHandlers) ProposalNew(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.RequireUser(w, r, h.sessions, h.users)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	p := domain.Proposal{
		AuthorID: user.ID,
		Status:   domain.ProposalStatusDraft,
	}
	if err := h.proposals.InsertProposal(r.Context(), &p); err != nil {
		http.Error(w, "failed to create draft", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/proposals/"+strconv.FormatInt(p.ID, 10)+"/edit", http.StatusSeeOther)
}

func (h *PageHandlers) ProposalEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.RequireUser(w, r, h.sessions, h.users)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	idStr := chi.URLParam(r, "id")
	pid, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	p, ok := h.proposals.GetProposalByID(r.Context(), pid)
	if !ok || p.AuthorID != user.ID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if p.Status != domain.ProposalStatusDraft && p.Status != domain.ProposalStatusChangesRequested {
		http.Redirect(w, r, "/proposals/"+strconv.FormatInt(pid, 10), http.StatusSeeOther)
		return
	}

	data := &TemplateData{
		User:     &user,
		Proposal: &p,
		Page:     "proposal_edit.html",
	}
	Render(w, data)
}

func (h *PageHandlers) ProposalView(w http.ResponseWriter, r *http.Request) {
	user, ok := actorFromRequest(r, h.sessions, h.users)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

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

	p, ok := h.proposals.GetProposalByID(r.Context(), pid)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if user.Role != domain.UserRoleAdmin && p.AuthorID != user.ID {
		http.NotFound(w, r)
		return
	}
	if user.Role == domain.UserRoleAdmin && p.Status != domain.ProposalStatusSubmitted {
		http.NotFound(w, r)
		return
	}

	proposalJSON, _ := json.Marshal(p)

	data := &TemplateData{
		User:         &user,
		Proposal:     &p,
		ProposalJSON: string(proposalJSON),
		Page:         "proposal_view.html",
	}
	Render(w, data)
}

func (h *PageHandlers) Profile(w http.ResponseWriter, r *http.Request) {
	user, ok := actorFromRequest(r, h.sessions, h.users)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := &TemplateData{User: &user, Page: "profile.html"}
	Render(w, data)
}
