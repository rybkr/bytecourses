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
	authH := handlers.NewAuthHandler(a.UserStore, a.SessionStore)
	utilH := handlers.NewUtilHandlers()
	propH := handlers.NewProposalHandlers(a.ProposalStore, a.UserStore, a.SessionStore)
	pageH := handlers.NewPageHandlers(a.UserStore, a.SessionStore, a.ProposalStore)

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.Logger)

	r.Post("/api/register", authH.Register)
	r.Post("/api/login", authH.Login)
	r.Post("/api/logout", authH.Logout)
	r.Get("/api/me", authH.Me)
	r.Get("/api/health", utilH.Health)

	r.Route("/api/proposals", func(r chi.Router) {
		r.Use(appmw.RequireUser(a.SessionStore, a.UserStore))

		r.Post("/", propH.Create)
		r.Get("/", propH.List)
		r.Get("/mine", propH.ListMine)

		r.Route("/{id}", func(r chi.Router) {
			r.Use(appmw.RequireProposal(a.ProposalStore, proposalID))

			r.Get("/", propH.Get)
			r.Patch("/", propH.Update)
			r.Delete("/", propH.Delete)

			r.Route("/actions", func(r chi.Router) {
				r.Post("/{action}", propH.Action)
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
