package handlers

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"

	"bytecourses/internal/domain"
	"bytecourses/internal/infrastructure/http/middleware"
	"bytecourses/internal/pkg/errors"
	"bytecourses/internal/services"
)

// PageData is the standard data structure passed to all page templates.
// Templates access User directly as .User and page-specific data via .Data
type PageData struct {
	User *domain.User
	Data any
}

// ProposalPageData is the data structure for proposal pages.
// Templates access User and Proposal directly at root level.
type ProposalPageData struct {
	User     *domain.User
	Proposal *domain.Proposal
}

// PageHandler handles rendering of HTML page templates.
type PageHandler struct {
	templates       map[string]*template.Template
	funcMap         template.FuncMap
	proposalService *services.ProposalService
}

// NewPageHandler creates a new PageHandler by parsing all page templates.
// Each page template is combined with the layout template.
func NewPageHandler(templatesDir string, proposalService *services.ProposalService) *PageHandler {
	funcMap := template.FuncMap{
		"markdown": renderMarkdown,
	}

	h := &PageHandler{
		templates:       make(map[string]*template.Template),
		funcMap:         funcMap,
		proposalService: proposalService,
	}

	layoutPath := filepath.Join(templatesDir, "layout.html")
	pagesDir := filepath.Join(templatesDir, "pages")

	// Read all page templates
	entries, err := os.ReadDir(pagesDir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		pagePath := filepath.Join(pagesDir, entry.Name())

		// Parse layout first, then page template
		// This allows page to define blocks that override layout defaults
		tmpl, err := template.New("").Funcs(funcMap).ParseFiles(layoutPath, pagePath)
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

// render executes a page template with the given data.
// It automatically includes the current user from the request context.
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute the page template which defines blocks and invokes the layout
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

// NotFound renders the 404 page.
func (h *PageHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	h.render(w, r, "404.html", nil)
}

// Home renders the home page.
func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "home.html", nil)
}

// Login renders the login page.
func (h *PageHandler) Login(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "login.html", nil)
}

// Register renders the registration page.
func (h *PageHandler) Register(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "register.html", nil)
}

// RequestPasswordReset renders the forgot password page.
func (h *PageHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "forgot_password.html", nil)
}

// ConfirmPasswordReset renders the reset password page.
func (h *PageHandler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "reset_password.html", nil)
}

// Profile renders the user profile page.
func (h *PageHandler) Profile(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "profile.html", nil)
}

// Courses renders the course listing page.
func (h *PageHandler) Courses(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "courses.html", nil)
}

// CourseView renders a single course page.
func (h *PageHandler) CourseView(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "course_view.html", nil)
}

// CourseEdit renders the course edit page.
func (h *PageHandler) CourseEdit(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "course_edit.html", nil)
}

// Proposals renders the proposals listing page.
func (h *PageHandler) Proposals(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "proposals.html", nil)
}

// ProposalView renders a single proposal page.
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "proposal_view.html", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

// ProposalEdit renders the proposal edit page.
func (h *PageHandler) ProposalEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if this is a new proposal (no ID in URL) or editing existing one
	proposalIDStr := chi.URLParam(r, "id")
	if proposalIDStr == "" {
		// New proposal - create empty proposal data
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

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var buf bytes.Buffer
		if err := tmpl.ExecuteTemplate(&buf, "proposal_edit.html", pd); err != nil {
			log.Printf("template execution error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "proposal_edit.html", pd); err != nil {
		log.Printf("template execution error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

// LectureView renders a single lecture page.
func (h *PageHandler) LectureView(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "lecture_view.html", nil)
}

// LectureEdit renders the lecture edit page.
func (h *PageHandler) LectureEdit(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, "lecture_edit.html", nil)
}
