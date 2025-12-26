package main

import (
	"bytecourses/internal/auth/memsession"
	"bytecourses/internal/http/handlers"
	"bytecourses/internal/store/memstore"
	"log"
	"net/http"
	"time"
)

func main() {
	userStore := memstore.NewUserStore()
    proposalStore := memstore.NewProposalStore()
    sessionStore := memsession.New(24 * time.Hour)

    authHandlers := handlers.NewAuthHandlers(userStore, sessionStore)
    utilHandlers := handlers.NewUtilHandlers()
    proposalHandlers := handlers.NewProposalHandlers(proposalStore, userStore, sessionStore)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/register", authHandlers.Register)
	mux.HandleFunc("/api/login", authHandlers.Login)
	mux.HandleFunc("/api/logout", authHandlers.Logout)
	mux.HandleFunc("/api/me", authHandlers.Me)

    mux.HandleFunc("/api/health", utilHandlers.Health)

    mux.HandleFunc("/api/proposals", proposalHandlers.Proposals)
    mux.HandleFunc("/api/proposals/", proposalHandlers.ProposalByID)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
