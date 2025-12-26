package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"encoding/json"
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseGlob("web/templates/**/*.html").Funcs(template.FuncMap{
	"safeJS": func(s string) template.JS {
		return template.JS(s)
	},
}))

type TemplateData struct {
	User         *domain.User
	Proposal     *domain.Proposal
	ProposalJSON string
	Page         string // Page template name to use
}

func Render(w http.ResponseWriter, data *TemplateData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Execute the page template, which will include the layout
	templateName := "pages/home" // default
	if data != nil && data.Page != "" {
		templateName = data.Page
	}
	
	if err := templates.ExecuteTemplate(w, templateName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RenderWithUser(w http.ResponseWriter, r *http.Request, sessions auth.SessionStore, users store.UserStore, data *TemplateData) {
	if data == nil {
		data = &TemplateData{}
	}
	
	// Try to get user from session
	if user, ok := actorFromRequest(r, sessions, users); ok {
		data.User = &user
	}
	
	Render(w, data)
}

