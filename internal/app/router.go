package app

import (
	"bytecourses/internal/http/handlers"
	appmw "bytecourses/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"net/http"
	"strconv"
)

func proposalID(r *http.Request) (int64, bool) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	return id, err == nil
}

func (a *App) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.Logger)

	authH := handlers.NewAuthHandler(a.UserStore, a.SessionStore)
	sysH := handlers.NewSystemHandlers(a.UserStore)
	propH := handlers.NewProposalHandlers(a.ProposalStore, a.UserStore, a.SessionStore)
	pageH := handlers.NewPageHandlers(a.UserStore, a.SessionStore, a.ProposalStore)

	r.Route("/api", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
		r.Post("/logout", authH.Logout)
		r.With(appmw.RequireUser(a.SessionStore, a.UserStore)).Get("/me", authH.Me)
		r.With(appmw.RequireUser(a.SessionStore, a.UserStore)).Patch("/profile", authH.UpdateProfile)
		r.Get("/health", sysH.Health)
		r.Get("/diagnostics", sysH.Diagnostics)

		r.Route("/proposals", func(r chi.Router) {
			r.Use(appmw.RequireUser(a.SessionStore, a.UserStore))

			r.Post("/", propH.Create)
			r.Get("/", propH.List)
			r.Get("/mine", propH.ListMine)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(appmw.RequireProposal(a.ProposalStore, proposalID))

				r.Get("/", propH.Get)
				r.Patch("/", propH.Update)
				r.Delete("/", propH.Delete)
				r.Post("/actions/{action}", propH.Action)
			})
		})
	})

	r.Get("/", pageH.Home)
	r.Get("/login", pageH.Login)
	r.Get("/register", pageH.Register)
	r.With(appmw.RequireLogin(a.SessionStore, a.UserStore)).Get("/profile", pageH.Profile)

	r.Route("/proposals", func(r chi.Router) {
		r.Use(appmw.RequireLogin(a.SessionStore, a.UserStore))
		r.Get("/", pageH.ProposalsList)
		r.Get("/mine", pageH.ProposalsListMine)
		r.Get("/new", pageH.ProposalNew)
		r.Get("/{id}", pageH.ProposalView)
		r.Get("/{id}/edit", pageH.ProposalEdit)
	})

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	return r
}
