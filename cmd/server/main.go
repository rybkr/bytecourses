package main

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/auth/memsession"
	"bytecourses/internal/domain"
	"bytecourses/internal/http/handlers"
	"bytecourses/internal/store"
	"bytecourses/internal/store/memstore"
	"context"
	"flag"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

func ensureTestAdmin(users store.UserStore) error {
	email := "admin@local.bytecourses.org"
	password := "admin"
	if _, ok := users.GetUserByEmail(context.Background(), email); ok {
		return nil
	}

	hash, _ := auth.HashPassword(password)
	return users.InsertUser(context.Background(), &domain.User{
		Email:        email,
		PasswordHash: hash,
		Role:         domain.UserRoleAdmin,
	})
}

func main() {
	addr := flag.String("addr", ":8080", "http listen address")
	seedAdmin := flag.Bool("seed-admin", false, "seed a test admin user")
	bcryptCost := flag.Int("bcrypt-cost", bcrypt.DefaultCost, "bcrypt cost factor")
	flag.Parse()

	userStore := memstore.NewUserStore()
	proposalStore := memstore.NewProposalStore()
	sessionStore := memsession.New(24 * time.Hour)

	auth.SetBcryptCost(*bcryptCost)
	if *seedAdmin {
		ensureTestAdmin(userStore)
	}

	authHandlers := handlers.NewAuthHandler(userStore, sessionStore)
	utilHandlers := handlers.NewUtilHandlers()
	proposalHandlers := handlers.NewProposalHandlers(proposalStore, userStore, sessionStore)
	pageHandlers := handlers.NewPageHandlers(userStore, sessionStore, proposalStore)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Post("/api/register", authHandlers.Register)
	r.Post("/api/login", authHandlers.Login)
	r.Post("/api/logout", authHandlers.Logout)
	r.Get("/api/me", authHandlers.Me)
	r.Get("/api/health", utilHandlers.Health)

	r.Route("/api/proposals", func(r chi.Router) {
		r.Post("/", proposalHandlers.Create)
		r.Get("/", proposalHandlers.ListMine)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", proposalHandlers.Get)
			r.Patch("/", proposalHandlers.Update)

			r.Route("/actions", func(r chi.Router) {
				r.Post("/{action}", proposalHandlers.Action) // submit/withdraw/etc
			})
		})
	})

	r.Get("/", pageHandlers.Home)
	r.Get("/login", pageHandlers.Login)
	r.Get("/register", pageHandlers.Register)
	r.Get("/profile", pageHandlers.Profile)
	r.Get("/proposals", pageHandlers.ProposalsList)
	r.Get("/proposals/new", pageHandlers.ProposalNew)
	r.Get("/proposals/{id}", pageHandlers.ProposalView)

	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, r))
}
