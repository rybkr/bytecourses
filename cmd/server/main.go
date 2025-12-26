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
	flag.Parse()

	userStore := memstore.NewUserStore()
	proposalStore := memstore.NewProposalStore()
	sessionStore := memsession.New(24 * time.Hour)

	if *seedAdmin {
		ensureTestAdmin(userStore)
	}

	authHandlers := handlers.NewAuthHandler(userStore, sessionStore)
	utilHandlers := handlers.NewUtilHandlers()
	proposalHandlers := handlers.NewProposalHandler(proposalStore, userStore, sessionStore)

    mux := http.NewServeMux()

	mux.HandleFunc("/api/register", authHandlers.Register)
	mux.HandleFunc("/api/login", authHandlers.Login)
	mux.HandleFunc("/api/logout", authHandlers.Logout)
	mux.HandleFunc("/api/me", authHandlers.Me)

	mux.HandleFunc("/api/health", utilHandlers.Health)

	mux.HandleFunc("/api/proposals", proposalHandlers.Proposals)
	mux.HandleFunc("/api/proposals/", proposalHandlers.ProposalByID)

	mux.Handle("/", http.FileServer(http.Dir("web")))

	log.Printf("listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
