package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"html/template"
	"net/http"
)

var templates = template.Must(
	template.New("").
		Funcs(template.FuncMap{
			"safeJS": func(s string) template.JS {
				return template.JS(s)
			},
		}).
		ParseGlob("web/templates/**/*.html"),
)

type TemplateData struct {
	User         *domain.User
	Proposal     *domain.Proposal
	ProposalJSON string
	Page         string // Page template name to use
}

func Render(w http.ResponseWriter, data *TemplateData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// Execute the page template, which will include the layout
	// ParseGlob names templates based on file paths
	templateName := "home" // default
	if data != nil && data.Page != "" {
		templateName = data.Page
	}
	
	// Try different possible template name formats
	// ParseGlob typically uses the file path relative to the pattern
	possibleNames := []string{
		"pages/" + templateName,
		templateName,
		"web/templates/pages/" + templateName,
	}
	
	var foundName string
	for _, name := range possibleNames {
		if templates.Lookup(name) != nil {
			foundName = name
			break
		}
	}
	
	if foundName == "" {
		foundName = templateName // fallback, will error if not found
	}
	
	if err := templates.ExecuteTemplate(w, foundName, data); err != nil {
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

