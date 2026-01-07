package handlers

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"bytes"
	"html/template"
	"net/http"
	"os"
	"sync"
)

var (
	layoutContent []byte
	layoutOnce    sync.Once
	funcMap       = template.FuncMap{
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
	}
)

func getLayoutContent() []byte {
	layoutOnce.Do(func() {
		var err error
		layoutContent, err = os.ReadFile("web/templates/layout.html")
		if err != nil {
			panic(err)
		}
	})
	return layoutContent
}

type TemplateData struct {
	User         *domain.User
	Proposal     *domain.Proposal
	ProposalJSON string
	Page         string
}

func Render(w http.ResponseWriter, data *TemplateData) {
	// Execute the page template, which will include the layout
	templateName := "home.html" // default
	if data != nil && data.Page != "" {
		// Add .html extension if not present
		if len(data.Page) < 5 || data.Page[len(data.Page)-5:] != ".html" {
			templateName = data.Page + ".html"
		} else {
			templateName = data.Page
		}
	}

	// Parse layout and page template together so blocks are scoped correctly
	layoutContent := getLayoutContent()
	templates := template.Must(template.New("layout").Funcs(funcMap).Parse(string(layoutContent)))
	templates = template.Must(templates.ParseFiles("web/templates/pages/" + templateName))

	// Execute template to a buffer first to catch errors before writing headers
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers and write the rendered template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}

func RenderWithUser(w http.ResponseWriter, r *http.Request, sessions auth.SessionStore, users store.UserStore, data *TemplateData) {
	if data == nil {
		data = &TemplateData{}
	}

	// Try to get user from session
	if user, ok := actorFromRequest(r, sessions, users); ok {
		data.User = user
	}

	Render(w, data)
}
