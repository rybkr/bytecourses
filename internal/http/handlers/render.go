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

	"github.com/yuin/goldmark"
)

var (
	layoutContent []byte
	layoutOnce    sync.Once
	markdown      = goldmark.New()
	funcMap       = template.FuncMap{
		"safeJS": func(s string) template.JS {
			return template.JS(s)
		},
		"markdown": func(s string) template.HTML {
			var buf bytes.Buffer
			if err := markdown.Convert([]byte(s), &buf); err != nil {
				return template.HTML("")
			}
			return template.HTML(buf.String())
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
	Course       *domain.Course
	CourseJSON   string
	Courses      []domain.Course
	Instructor   *domain.User
	Page         string
}

func Render(w http.ResponseWriter, data *TemplateData) {
	templateName := "home.html"
	if data != nil && data.Page != "" {
		if len(data.Page) < 5 || data.Page[len(data.Page)-5:] != ".html" {
			templateName = data.Page + ".html"
		} else {
			templateName = data.Page
		}
	}

	layoutContent := getLayoutContent()
	templates := template.Must(template.New("layout").Funcs(funcMap).Parse(string(layoutContent)))
	templates = template.Must(templates.ParseFiles("web/templates/pages/" + templateName))

	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	buf.WriteTo(w)
}

func RenderWithUser(w http.ResponseWriter, r *http.Request, sessions auth.SessionStore, users store.UserStore, data *TemplateData) {
	if data == nil {
		data = &TemplateData{}
	}

	c, err := r.Cookie("session")
	if err == nil && c.Value != "" {
		if uid, ok := sessions.GetUserIDByToken(c.Value); ok {
			if user, ok := users.GetUserByID(r.Context(), uid); ok {
				data.User = user
			}
		}
	}

	Render(w, data)
}
