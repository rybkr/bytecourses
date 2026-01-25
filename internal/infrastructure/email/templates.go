package email

import (
	"embed"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	welcomeTemplate           *template.Template
	passwordResetTemplate     *template.Template
	proposalSubmittedTemplate *template.Template
	proposalApprovedTemplate  *template.Template
	proposalRejectedTemplate  *template.Template
	proposalChangesTemplate   *template.Template
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

	proposalSubmittedTemplate, err = template.ParseFS(templateFS, "templates/proposal_submitted.html")
	if err != nil {
		panic("failed to parse proposal submitted template: " + err.Error())
	}

	proposalApprovedTemplate, err = template.ParseFS(templateFS, "templates/proposal_approved.html")
	if err != nil {
		panic("failed to parse proposal approved template: " + err.Error())
	}

	proposalRejectedTemplate, err = template.ParseFS(templateFS, "templates/proposal_rejected.html")
	if err != nil {
		panic("failed to parse proposal rejected template: " + err.Error())
	}

	proposalChangesTemplate, err = template.ParseFS(templateFS, "templates/proposal_changes_requested.html")
	if err != nil {
		panic("failed to parse proposal changes requested template: " + err.Error())
	}
}
