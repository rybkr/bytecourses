package handlers

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/infrastructure/persistence"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/services"
)

type PageData struct {
	User *domain.User
	Data any
}

type ProposalPageData struct {
	User     *domain.User
	Proposal *domain.Proposal
}

type CoursesPageData struct {
	User         *domain.User
	Courses      []domain.Course
	Instructors  map[int64]*domain.User
	ModuleCounts map[int64]int
}

type PageHandler struct {
	templates       map[string]*template.Template
	funcMap         template.FuncMap
	proposalService *services.ProposalService
	courseService   *services.CourseService
	userRepo        persistence.UserRepository
}

func NewPageHandler(templatesFS embed.FS, proposalService *services.ProposalService, courseService *services.CourseService, userRepo persistence.UserRepository) *PageHandler {
	funcMap := template.FuncMap{
		"markdown": renderMarkdown,
	}

	h := &PageHandler{
		templates:       make(map[string]*template.Template),
		funcMap:         funcMap,
		proposalService: proposalService,
		courseService:   courseService,
		userRepo:        userRepo,
	}

	layoutContent, err := fs.ReadFile(templatesFS, "templates/layout.html")
	if err != nil {
		return nil
	}

	entries, err := fs.ReadDir(templatesFS, "templates/pages")
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		pageContent, err := fs.ReadFile(templatesFS, "templates/pages/"+entry.Name())
		if err != nil {
			return nil
		}

		tmpl, err := template.New("").Funcs(funcMap).Parse(string(layoutContent))
		if err != nil {
			return nil
		}

		tmpl, err = tmpl.Parse(string(pageContent))
		if err != nil {
			return nil
		}

		h.templates[entry.Name()] = tmpl
	}

	return h
}

func renderMarkdown(s string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(s), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(s))
	}
	return template.HTML(buf.String())
}

func (h *PageHandler) render(w http.ResponseWriter, r *http.Request, name string, data any) {
	tmpl, ok := h.templates[name]
	if !ok {
		log.Printf("template not found: %s", name)
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	user, _ := middleware.UserFromContext(r.Context())

	pd := PageData{
		User: user,
		Data: data,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	h.render(w, r, "404.html", nil)
}

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "home.html", nil)
}

func (h *PageHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "login.html", nil)
}

func (h *PageHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "register.html", nil)
}

func (h *PageHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "forgot_password.html", nil)
}

func (h *PageHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "reset_password.html", nil)
}

func (h *PageHandler) Profile(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "profile.html", nil)
}

func (h *PageHandler) Courses(w http.ResponseWriter, r *http.Request) {
	courses, err := h.courseService.List(r.Context())
	if err != nil {
		log.Printf("error fetching courses: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	instructors := make(map[int64]*domain.User)
	instructorIDs := make(map[int64]bool)
	for _, course := range courses {
		if !instructorIDs[course.InstructorID] {
			instructorIDs[course.InstructorID] = true
			if instructor, ok := h.userRepo.GetByID(r.Context(), course.InstructorID); ok {
				instructors[course.InstructorID] = instructor
			}
		}
	}

	user, _ := middleware.UserFromContext(r.Context())

	pd := CoursesPageData{
		User:         user,
		Courses:      courses,
		Instructors:  instructors,
		ModuleCounts: make(map[int64]int),
	}

	tmpl, ok := h.templates["courses.html"]
	if !ok {
		log.Printf("template not found: courses.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) CourseView(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "course_view.html", nil)
}

func (h *PageHandler) CourseEdit(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "course_edit.html", nil)
}

func (h *PageHandler) Proposals(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "proposals.html", nil)
}

func (h *PageHandler) ProposalView(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	proposalID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	proposal, err := h.proposalService.Get(r.Context(), &services.GetProposalQuery{
		ProposalID: proposalID,
		UserID:     user.ID,
		UserRole:   user.Role,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "proposal not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching proposal: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, ok := h.templates["proposal_view.html"]
	if !ok {
		log.Printf("template not found: proposal_view.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	pd := ProposalPageData{
		User:     user,
		Proposal: proposal,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) ProposalEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	proposalIDStr := chi.URLParam(r, "id")
	if proposalIDStr == "" {
		tmpl, ok := h.templates["proposal_edit.html"]
		if !ok {
			log.Printf("template not found: proposal_edit.html")
			http.Error(w, "page not found", http.StatusNotFound)
			return
		}

		pd := ProposalPageData{
			User: user,
			Proposal: &domain.Proposal{
				ID:     0,
				Status: domain.ProposalStatusDraft,
			},
		}

		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
			log.Printf("template execution error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		buf.WriteTo(w)
		return
	}

	proposalID, err := strconv.ParseInt(proposalIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	proposal, err := h.proposalService.Get(r.Context(), &services.GetProposalQuery{
		ProposalID: proposalID,
		UserID:     user.ID,
		UserRole:   user.Role,
	})
	if err != nil {
		if err == errors.ErrNotFound {
			http.Error(w, "proposal not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching proposal: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, ok := h.templates["proposal_edit.html"]
	if !ok {
		log.Printf("template not found: proposal_edit.html")
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	pd := ProposalPageData{
		User:     user,
		Proposal: proposal,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

func (h *PageHandler) LectureView(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "lecture_view.html", nil)
}

func (h *PageHandler) LectureEdit(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "lecture_edit.html", nil)
}
