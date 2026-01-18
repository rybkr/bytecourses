package http

import (
	"bytecourses/internal/bootstrap"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func NewRouter(c *bootstrap.Container) http.Handler {
    r := chi.NewRouter()
    r.Use(middleware.Recoverer)
    r.Use(middleware.Logger)
}
