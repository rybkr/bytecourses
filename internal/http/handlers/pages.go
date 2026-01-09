package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/services"
	"bytecourses/internal/store"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type PageHandlers struct {
	services  *services.Services
	users     store.UserStore
	sessions  auth.SessionStore
	proposals store.ProposalStore
}

func NewPageHandlers(services *services.Services, users store.UserStore, sessions auth.SessionStore, proposals store.ProposalStore) *PageHandlers {
	return &PageHandlers{
		services:  services,
		users:     users,
		sessions:  sessions,
		proposals: proposals,
	}
}

func (h *PageHandlers) Home(w http.ResponseWriter, r *http.Request) {
	if !requirePath(w, r, "/") {
		return
	}
	data := &TemplateData{
		Page: "home.html",
	}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) Login(w http.ResponseWriter, r *http.Request) {
	if _, ok := userFromRequest(r); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := &TemplateData{
		Page: "login.html",
	}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) Register(w http.ResponseWriter, r *http.Request) {
	if _, ok := userFromRequest(r); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := &TemplateData{
		Page: "register.html",
	}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if _, ok := userFromRequest(r); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := &TemplateData{
		Page: "forgot_password.html",
	}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	if _, ok := userFromRequest(r); ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := &TemplateData{
		Page: "reset_password.html",
	}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) ProposalsList(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromRequest(r)
	if !ok {
		return
	}
	if !u.IsAdmin() {
		http.Redirect(w, r, "/proposals/mine", http.StatusSeeOther)
		return
	}

	data := &TemplateData{
		User: u,
		Page: "proposals.html",
	}
	Render(w, data)
}

func (h *PageHandlers) ProposalsListMine(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromRequest(r)
	if !ok {
		return
	}

	data := &TemplateData{
		User: u,
		Page: "proposals.html",
	}
	Render(w, data)
}

func (h *PageHandlers) ProposalNew(w http.ResponseWriter, r *http.Request) {
	u, ok := userFromRequest(r)
	if !ok {
		return
	}

	proposal, err := h.services.Proposals.CreateProposal(r.Context(), &services.CreateProposalRequest{
		AuthorID: u.ID,
		// Empty fields - proposal starts as draft
	})
	if err != nil {
		http.Error(w, "failed to create draft", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/proposals/"+strconv.FormatInt(proposal.ID, 10)+"/edit", http.StatusSeeOther)
}

func (h *PageHandlers) ProposalEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromRequest(r)
	if !ok {
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
		User:     user,
		Proposal: p,
		Page:     "proposal_edit.html",
	}
	Render(w, data)
}

func (h *PageHandlers) ProposalView(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromRequest(r)
	if !ok {
		return
	}

	pidStr := chi.URLParam(r, "id")
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
	if !p.IsViewableBy(user) {
		http.NotFound(w, r)
		return
	}

	proposalJSON, _ := json.Marshal(p)

	data := &TemplateData{
		User:         user,
		Proposal:     p,
		ProposalJSON: string(proposalJSON),
		Page:         "proposal_view.html",
	}
	Render(w, data)
}

func (h *PageHandlers) Profile(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromRequest(r)
	if !ok {
		return
	}

	data := &TemplateData{User: user, Page: "profile.html"}
	Render(w, data)
}
