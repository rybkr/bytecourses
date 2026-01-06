package app

import (
	"bytecourses/internal/http/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func (a *App) Router() http.Handler {
    authH := handlers.NewAuthHandler(a.UserStore, a.SessionStore)
	utilH := handlers.NewUtilHandlers()
	propH := handlers.NewProposalHandlers(a.ProposalStore, a.UserStore, a.SessionStore)
	pageH := handlers.NewPageHandlers(a.UserStore, a.SessionStore, a.ProposalStore)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Post("/api/register", authH.Register)
	r.Post("/api/login", authH.Login)
	r.Post("/api/logout", authH.Logout)
	r.Get("/api/me", authH.Me)
	r.Get("/api/health", utilH.Health)

	r.Route("/api/proposals", func(r chi.Router) {
		r.With(propH.WithUser).Post("/", propH.Create)
		r.With(propH.WithUser).Get("/", propH.List)
		r.With(propH.WithUser).Get("/mine", propH.ListMine)

		r.Route("/{id}", func(r chi.Router) {
			r.With(propH.WithUser, propH.WithProposal).Get("/", propH.Get)
			r.With(propH.WithUser, propH.WithProposal).Patch("/", propH.Update)
			r.With(propH.WithUser, propH.WithProposal).Delete("/", propH.Delete)
			r.Route("/actions", func(r chi.Router) {
				r.With(propH.WithUser, propH.WithProposal).
					Post("/{action}", propH.Action)
			})
		})
	})

	r.Get("/", pageH.Home)
	r.Get("/login", pageH.Login)
	r.Get("/register", pageH.Register)
	r.Get("/profile", pageH.Profile)
	r.Get("/proposals", pageH.ProposalsList)
	r.Get("/proposals/mine", pageH.ProposalsListMine)
	r.Get("/proposals/new", pageH.ProposalNew)
	r.Get("/proposals/{id}", pageH.ProposalView)
	r.Get("/proposals/{id}/edit", pageH.ProposalEdit)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	return r
}
