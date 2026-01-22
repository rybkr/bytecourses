package email

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	welcomeTemplate *template.Template
    passwordResetTemplate *template.Template
)

func init() {
	var err error

	welcomeTemplate, err = template.ParseFS(templateFS, "templates/welcome.html")
	if err != nil {
		panic("failed to parse welcome template: " + err.Error())
	}

    passwordResetTemplate, err = template.ParseFS(templateFS, "templates/password_reset.html")
    if err != nil {
        panic("failed to parse password reset template: " + err.Error())
    }
}
