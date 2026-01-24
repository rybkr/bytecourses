package http

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"bytecourses/internal/bootstrap"
	"bytecourses/internal/infrastructure/http/handlers"
	"bytecourses/internal/infrastructure/http/middleware"
)

func NewRouter(c *bootstrap.Container, webFS embed.FS) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.Logger)

	pageHandler := handlers.NewPageHandler(webFS, c.ProposalService, c.CourseService, c.ModuleService, c.ContentService, c.UserRepo)
	authHandler := handlers.NewAuthHandler(c.AuthService)
	proposalHandler := handlers.NewProposalHandler(c.ProposalService, c.CourseService)
	courseHandler := handlers.NewCourseHandler(c.CourseService)
	moduleHandler := handlers.NewModuleHandler(c.ModuleService)
	contentHandler := handlers.NewContentHandler(c.ContentService)

	requireUser := middleware.RequireUser(c.SessionStore, c.UserRepo)
	requireAdmin := middleware.RequireAdmin(c.SessionStore, c.UserRepo)
	optionalUser := middleware.OptionalUser(c.SessionStore, c.UserRepo)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/password-reset/request", authHandler.RequestPasswordReset)
		r.Post("/password-reset/confirm", authHandler.ConfirmPasswordReset)

		r.With(requireUser).Post("/logout", authHandler.Logout)
		r.With(requireUser).Get("/me", authHandler.Me)
		r.With(requireUser).Patch("/me", authHandler.UpdateProfile)

		r.Route("/proposals", func(r chi.Router) {
			r.Use(requireUser)
			r.Post("/", proposalHandler.Create)
			r.Get("/", proposalHandler.List)
			r.Patch("/{id}", proposalHandler.Update)
			r.Delete("/{id}", proposalHandler.Delete)
			r.Get("/{id}", proposalHandler.Get)
			r.Post("/{id}/actions/submit", proposalHandler.Submit)
			r.Post("/{id}/actions/withdraw", proposalHandler.Withdraw)
			r.With(requireAdmin).Post("/{id}/actions/approve", proposalHandler.Approve)
			r.With(requireAdmin).Post("/{id}/actions/reject", proposalHandler.Reject)
			r.With(requireAdmin).Post("/{id}/actions/request-changes", proposalHandler.RequestChanges)
			r.Post("/{id}/actions/create-course", proposalHandler.CreateCourse)
		})

		r.Route("/courses", func(r chi.Router) {
			r.With(optionalUser).Get("/", courseHandler.List)

			r.With(requireUser).Get("/{id}", courseHandler.Get)
			r.With(requireUser).Patch("/{id}", courseHandler.Update)
			r.With(requireUser).Post("/{id}/publish", courseHandler.Publish)

			r.Route("/{courseId}/modules", func(r chi.Router) {
				r.Use(requireUser)
				r.Post("/", moduleHandler.Create)
				r.Get("/", moduleHandler.List)
				r.Get("/{moduleId}", moduleHandler.Get)
				r.Patch("/{moduleId}", moduleHandler.Update)
				r.Delete("/{moduleId}", moduleHandler.Delete)
				r.Post("/{moduleId}/publish", moduleHandler.Publish)
			})
		})

		r.Route("/modules/{moduleId}/readings", func(r chi.Router) {
			r.Use(requireUser)
			r.Post("/", contentHandler.CreateReading)
			r.Get("/", contentHandler.ListReadings)
			r.Get("/{readingId}", contentHandler.GetReading)
			r.Patch("/{readingId}", contentHandler.UpdateReading)
			r.Delete("/{readingId}", contentHandler.DeleteReading)
			r.Post("/{readingId}/publish", contentHandler.PublishReading)
		})
	})

	staticFS, _ := fs.Sub(webFS, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	r.Group(func(r chi.Router) {
		r.Use(optionalUser)
		r.Get("/", pageHandler.Home)
		r.Get("/login", pageHandler.Login)
		r.Get("/register", pageHandler.Register)
		r.Get("/forgot-password", pageHandler.RequestPasswordReset)
		r.Get("/reset-password", pageHandler.ConfirmPasswordReset)
		r.Get("/courses", pageHandler.Courses)
		r.Get("/courses/{id}", pageHandler.CourseView)
		r.Get("/courses/{id}/content", pageHandler.CourseContent)
	})

	r.Group(func(r chi.Router) {
		r.Use(requireUser)
		r.Get("/profile", pageHandler.Profile)

		r.Get("/proposals", pageHandler.Proposals)
		r.Get("/proposals/new", pageHandler.ProposalEdit)
		r.Get("/proposals/mine", pageHandler.Proposals)
		r.Get("/proposals/{id}", pageHandler.ProposalView)
		r.Get("/proposals/{id}/edit", pageHandler.ProposalEdit)

		r.Get("/courses/{id}/edit", pageHandler.CourseEdit)

		r.Get("/modules/{moduleId}/readings/{readingId}", pageHandler.LectureView)
		r.Get("/modules/{moduleId}/readings/{readingId}/edit", pageHandler.LectureEdit)
	})

	r.NotFound(pageHandler.NotFound)

	return r
}
