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
	courses   store.CourseStore
	modules   store.ModuleStore
	content   store.ContentStore
}

func NewPageHandlers(services *services.Services, users store.UserStore, sessions auth.SessionStore, proposals store.ProposalStore, courses store.CourseStore, modules store.ModuleStore, content store.ContentStore) *PageHandlers {
	return &PageHandlers{
		services:  services,
		users:     users,
		sessions:  sessions,
		proposals: proposals,
		courses:   courses,
		modules:   modules,
		content:   content,
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

func (h *PageHandlers) CoursesList(w http.ResponseWriter, r *http.Request) {
	courses, err := h.services.Courses.ListCourses(r.Context())
	if err != nil {
		http.Error(w, "failed to load courses", http.StatusInternalServerError)
		return
	}

	instructors := make(map[int64]*domain.User)
	moduleCounts := make(map[int64]int)

	for _, c := range courses {
		if _, exists := instructors[c.InstructorID]; !exists {
			if instructor, ok := h.users.GetUserByID(r.Context(), c.InstructorID); ok {
				instructors[c.InstructorID] = instructor
			}
		}

		if modules, err := h.modules.ListModulesByCourseID(r.Context(), c.ID); err == nil {
			moduleCounts[c.ID] = len(modules)
		}
	}

	data := &TemplateData{
		Courses:      courses,
		Instructors:  instructors,
		ModuleCounts: moduleCounts,
		Page:         "courses.html",
	}
	RenderWithUser(w, r, h.sessions, h.users, data)
}

func (h *PageHandlers) CourseView(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromRequest(r)
	if !ok {
		return
	}

	c, ok := courseFromRequest(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	course, err := h.services.Courses.GetCourse(r.Context(), c, user)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var instructor *domain.User
	if instructorUser, ok := h.users.GetUserByID(r.Context(), course.InstructorID); ok {
		instructor = instructorUser
	}

	courseJSON, _ := json.Marshal(course)

	var modules []domain.Module
	if courseModules, err := h.services.Modules.ListModules(r.Context(), course, user); err == nil {
		modules = courseModules
	}

	data := &TemplateData{
		User:         user,
		Course:       course,
		CourseJSON:   string(courseJSON),
		Instructor:   instructor,
		Modules:      modules,
		IsInstructor: course.IsTaughtBy(user),
		Page:         "course_view.html",
	}
	Render(w, data)
}

func (h *PageHandlers) CourseEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromRequest(r)
	if !ok {
		return
	}

	idStr := chi.URLParam(r, "id")
	cid, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	c, ok := h.courses.GetCourseByID(r.Context(), cid)
	if !ok || c.InstructorID != user.ID {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	if c.Status != domain.CourseStatusDraft {
		http.Redirect(w, r, "/courses/"+strconv.FormatInt(cid, 10), http.StatusSeeOther)
		return
	}

	courseJSON, _ := json.Marshal(c)

	data := &TemplateData{
		User:       user,
		Course:     c,
		CourseJSON: string(courseJSON),
		Page:       "course_edit.html",
	}
	Render(w, data)
}

func (h *PageHandlers) LectureEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromRequest(r)
	if !ok {
		return
	}

	course, ok := courseFromRequest(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	module, ok := moduleFromRequest(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	item, ok := contentItemFromRequest(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Verify instructor access
	if !course.IsTaughtBy(user) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Verify ownership chain
	if module.CourseID != course.ID || item.ModuleID != module.ID {
		http.NotFound(w, r)
		return
	}

	lecture, _ := h.content.GetLecture(r.Context(), item.ID)
	if lecture == nil {
		lecture = &domain.Lecture{
			ContentItemID: item.ID,
			Content:       "",
		}
	}

	data := &TemplateData{
		User:         user,
		Course:       course,
		Module:       module,
		ContentItem:  item,
		Lecture:      lecture,
		IsInstructor: true,
		Page:         "lecture_edit.html",
	}
	Render(w, data)
}

func (h *PageHandlers) LectureView(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromRequest(r)
	if !ok {
		return
	}

	course, ok := courseFromRequest(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	module, ok := moduleFromRequest(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	item, ok := contentItemFromRequest(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Verify ownership chain
	if module.CourseID != course.ID || item.ModuleID != module.ID {
		http.NotFound(w, r)
		return
	}

	// Non-instructors can only see published content
	if !course.IsTaughtBy(user) && item.Status != domain.ContentStatusPublished {
		http.NotFound(w, r)
		return
	}

	lecture, _ := h.content.GetLecture(r.Context(), item.ID)
	if lecture == nil {
		lecture = &domain.Lecture{
			ContentItemID: item.ID,
			Content:       "",
		}
	}

	data := &TemplateData{
		User:         user,
		Course:       course,
		Module:       module,
		ContentItem:  item,
		Lecture:      lecture,
		IsInstructor: course.IsTaughtBy(user),
		Page:         "lecture_view.html",
	}
	Render(w, data)
}

func (h *PageHandlers) NotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	Render(w, &TemplateData{
		Page: "404.html",
	})
}
