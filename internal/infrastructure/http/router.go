package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"bytecourses/internal/bootstrap"
	"bytecourses/internal/infrastructure/http/handlers"
	"bytecourses/internal/infrastructure/http/middleware"
)

func NewRouter(c *bootstrap.Container, templatesDir string) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.Logger)

	pageHandler := handlers.NewPageHandler(templatesDir)
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
			r.Get("/", proposalHandler.ListMine)
			r.Get("/{id}", proposalHandler.GetByID)
			r.Patch("/{id}", proposalHandler.Update)
			r.Delete("/{id}", proposalHandler.Delete)
			r.Post("/{id}/actions/submit", proposalHandler.Submit)
			r.Post("/{id}/actions/withdraw", proposalHandler.Withdraw)
		})

		r.With(requireAdmin).Get("/admin/proposals", proposalHandler.ListAll)
		r.With(requireAdmin).Post("/proposals/{id}/actions/approve", proposalHandler.Approve)
		r.With(requireAdmin).Post("/proposals/{id}/actions/reject", proposalHandler.Reject)
		r.With(requireAdmin).Post("/proposals/{id}/actions/request-changes", proposalHandler.RequestChanges)

		r.Route("/courses", func(r chi.Router) {
			r.With(optionalUser).Get("/", courseHandler.ListLive)

			r.With(requireUser).Post("/", courseHandler.Create)
			r.With(requireUser).Post("/from-proposal", courseHandler.CreateFromProposal)
			r.With(requireUser).Get("/{id}", courseHandler.GetByID)
			r.With(requireUser).Patch("/{id}", courseHandler.Update)
			r.With(requireUser).Post("/{id}/publish", courseHandler.Publish)
		})
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

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
