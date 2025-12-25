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
	_ = memsession.New(24 * time.Hour)

	register := handlers.NewRegisterHandler(userStore)
	login := handlers.NewLoginHandler(userStore, sessionStore)
    logout := handlers.NewLogoutHandler(sessionStore)

	mux := http.NewServeMux()
	mux.Handle("/api/register", register)
	mux.Handle("/api/login", login)
    mux.Handle("/api/logout", logout)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
