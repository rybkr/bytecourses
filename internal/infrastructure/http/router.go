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

	pageHandler := handlers.NewPageHandler(webFS, c.ProposalService, c.CourseService, c.UserRepo)
	authHandler := handlers.NewAuthHandler(c.AuthService)
	proposalHandler := handlers.NewProposalHandler(c.ProposalService)
	courseHandler := handlers.NewCourseHandler(c.CourseService)

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
		})

		r.Route("/courses", func(r chi.Router) {
			r.With(optionalUser).Get("/", courseHandler.List)

			r.With(requireUser).Post("/", courseHandler.Create)
			r.With(requireUser).Get("/{id}", courseHandler.Get)
			r.With(requireUser).Patch("/{id}", courseHandler.Update)
			r.With(requireUser).Post("/{id}/publish", courseHandler.Publish)
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

		r.Get("/lectures/{id}", pageHandler.LectureView)
		r.Get("/lectures/{id}/edit", pageHandler.LectureEdit)
	})

	r.NotFound(pageHandler.NotFound)

	return r
}
