package main

import (
	"github.com/rybkr/bytecourses/internal/handlers"
	"github.com/rybkr/bytecourses/internal/middleware"
	"github.com/rybkr/bytecourses/internal/store"
	"log"
	"net/http"
	"os"
)

func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://localhost/bytecourses?sslmode=disable"
		log.Println("using default database connection string")
	}

	store, err := store.New(connString)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer store.Close()

	authHandler := handlers.NewAuthHandler(store)
	courseHandler := handlers.NewCourseHandler(store)
	adminHandler := handlers.NewAdminHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/signup", authHandler.Signup)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)

	mux.HandleFunc("POST /api/courses", middleware.Auth(courseHandler.CreateCourse))
	mux.HandleFunc("GET /api/courses", courseHandler.ListCourses)
	mux.HandleFunc("DELETE /api/courses", middleware.Auth(courseHandler.DeleteCourse))

	mux.HandleFunc("GET /api/admin/users", middleware.RequireAdmin(adminHandler.ListUsers))
	mux.HandleFunc("GET /api/admin/courses", middleware.RequireAdmin(courseHandler.ListCourses))
	mux.HandleFunc("PATCH /api/admin/courses/approve", middleware.RequireAdmin(adminHandler.ApproveCourse))
	mux.HandleFunc("PATCH /api/admin/courses/reject", middleware.RequireAdmin(adminHandler.RejectCourse))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			log.Printf("404 not found: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})

	addr := ":8080"
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
