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

func courseID(r *http.Request) (int64, bool) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	return id, err == nil
}

func moduleID(r *http.Request) (int64, bool) {
	idStr := chi.URLParam(r, "moduleId")
	if idStr == "" {
		return 0, false
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	return id, err == nil
}

func contentID(r *http.Request) (int64, bool) {
	idStr := chi.URLParam(r, "contentId")
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

	authH := handlers.NewAuthHandler(a.Services)
	sysH := handlers.NewSystemHandlers(a.DB)
	propH := handlers.NewProposalHandler(a.Services)
	pageH := handlers.NewPageHandlers(a.Services, a.UserStore, a.SessionStore, a.ProposalStore, a.CourseStore, a.ModuleStore, a.ContentStore)
	courseH := handlers.NewCourseHandler(a.Services)
	moduleH := handlers.NewModuleHandler(a.Services)
	contentH := handlers.NewContentHandler(a.Services)

	r.Route("/api", func(r chi.Router) {
		r.Post("/register", authH.Register)
		r.Post("/login", authH.Login)
		r.Post("/logout", authH.Logout)

		r.Post("/password-reset/request", authH.RequestPasswordReset)
		r.Post("/password-reset/confirm", authH.ConfirmPasswordReset)
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

		r.Route("/courses", func(r chi.Router) {
			r.With(appmw.RequireUser(a.SessionStore, a.UserStore)).Post("/", courseH.Create)
			r.Get("/", courseH.List)

			r.Route("/{id}", func(r chi.Router) {
				r.Use(appmw.RequireUser(a.SessionStore, a.UserStore))
				r.Use(appmw.RequireCourse(a.CourseStore, courseID))
				r.Get("/", courseH.Get)
				r.Patch("/", courseH.Update)
				r.Post("/actions/{action}", courseH.Action)

				r.Route("/modules", func(r chi.Router) {
					r.Post("/", moduleH.Create)
					r.Get("/", moduleH.List)
					r.Post("/reorder", moduleH.Reorder)

					r.Route("/{moduleId}", func(r chi.Router) {
						r.Use(appmw.RequireModule(a.ModuleStore, moduleID))
						r.Get("/", moduleH.Get)
						r.Patch("/", moduleH.Update)
						r.Delete("/", moduleH.Delete)

						r.Route("/content", func(r chi.Router) {
							r.Post("/", contentH.CreateLecture)
							r.Get("/", contentH.ListContent)
							r.Post("/reorder", contentH.ReorderContent)

							r.Route("/{contentId}", func(r chi.Router) {
								r.Use(appmw.RequireContentItem(a.ContentStore, contentID))
								r.Get("/", contentH.GetLecture)
								r.Patch("/", contentH.UpdateLecture)
								r.Delete("/", contentH.DeleteContent)
								r.Post("/publish", contentH.PublishContent)
								r.Post("/unpublish", contentH.UnpublishContent)
							})
						})
					})
				})
			})
		})
	})

	r.Get("/", pageH.Home)
	r.Get("/login", pageH.Login)
	r.Get("/register", pageH.Register)
	r.Get("/forgot-password", pageH.RequestPasswordReset)
	r.Get("/reset-password", pageH.ConfirmPasswordReset)
	r.With(appmw.RequireLogin(a.SessionStore, a.UserStore)).Get("/profile", pageH.Profile)

	r.Route("/proposals", func(r chi.Router) {
		r.Use(appmw.RequireLogin(a.SessionStore, a.UserStore))
		r.Get("/", pageH.ProposalsList)
		r.Get("/mine", pageH.ProposalsListMine)
		r.Get("/new", pageH.ProposalNew)
		r.Get("/{id}", pageH.ProposalView)
		r.Get("/{id}/edit", pageH.ProposalEdit)
	})

	r.Get("/courses", pageH.CoursesList)
	r.Route("/courses/{id}", func(r chi.Router) {
		r.Use(appmw.RequireLogin(a.SessionStore, a.UserStore))
		r.Use(appmw.RequireCourse(a.CourseStore, courseID))
		r.Get("/", pageH.CourseView)
		r.Get("/edit", pageH.CourseEdit)

		r.Route("/modules/{moduleId}/content/{contentId}", func(r chi.Router) {
			r.Use(appmw.RequireModule(a.ModuleStore, moduleID))
			r.Use(appmw.RequireContentItem(a.ContentStore, contentID))
			r.Get("/", pageH.LectureView)
			r.Get("/edit", pageH.LectureEdit)
		})
	})

	r.NotFound(pageH.NotFound)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	return r
}
