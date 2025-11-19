package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rybkr/bytecourses/internal/handlers"
	"github.com/rybkr/bytecourses/internal/middleware"
	"github.com/rybkr/bytecourses/internal/store"
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
	instructorHandler := handlers.NewInstructorHandler(store)
	userHandler := handlers.NewUserHandler(store)
	applicationHandler := handlers.NewApplicationHandler(store)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/auth/signup", authHandler.Signup)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)

	mux.HandleFunc("GET /api/profile", middleware.Auth(userHandler.GetProfile))
	mux.HandleFunc("PATCH /api/profile", middleware.Auth(userHandler.UpdateProfile))
	mux.HandleFunc("GET /api/users", userHandler.GetUserByID)

	// Application routes
	mux.HandleFunc("POST /api/applications", middleware.Auth(applicationHandler.CreateApplication))
	mux.HandleFunc("GET /api/instructor/applications", middleware.Auth(applicationHandler.GetMyApplications))
	mux.HandleFunc("PATCH /api/instructor/applications", middleware.Auth(applicationHandler.UpdateApplication))
	mux.HandleFunc("DELETE /api/instructor/applications", middleware.Auth(applicationHandler.DeleteApplication))
	mux.HandleFunc("PATCH /api/instructor/applications/submit", middleware.Auth(applicationHandler.SubmitApplication))

	// Course routes (browsing only - all courses are approved)
	mux.HandleFunc("GET /api/courses/{id}", courseHandler.GetCourse)
	mux.HandleFunc("GET /api/courses", courseHandler.ListCourses)

	// Instructor course routes (published courses only)
	mux.HandleFunc("GET /api/instructor/courses", middleware.Auth(instructorHandler.GetMyCourses))
	mux.HandleFunc("PATCH /api/instructor/courses", middleware.Auth(instructorHandler.UpdateCourse))
	mux.HandleFunc("DELETE /api/instructor/courses", middleware.Auth(instructorHandler.DeleteCourse))

	// Admin routes
	mux.HandleFunc("GET /api/admin/users", middleware.RequireAdmin(adminHandler.ListUsers))
	mux.HandleFunc("GET /api/admin/applications", middleware.RequireAdmin(adminHandler.ListPendingApplications))
	mux.HandleFunc("PATCH /api/admin/applications/approve", middleware.RequireAdmin(adminHandler.ApproveApplication))
	mux.HandleFunc("PATCH /api/admin/applications/reject", middleware.RequireAdmin(adminHandler.RejectApplication))

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/course/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/course/index.html")
	})
	mux.HandleFunc("/about/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/about/index.html")
	})
	mux.HandleFunc("/profile/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/profile/index.html")
	})
	mux.HandleFunc("/apply/new/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/apply/new/index.html")
	})
	mux.HandleFunc("/apply/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/apply/index.html")
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			log.Printf("404 not found: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})

	// Start TTL cleanup goroutine for rejected applications
	go func() {
		ttlDays := 90
		if envTTL := os.Getenv("REJECTED_APPLICATION_TTL_DAYS"); envTTL != "" {
			if parsed, err := strconv.Atoi(envTTL); err == nil {
				ttlDays = parsed
			}
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		// Run immediately on startup
		ctx := context.Background()
		if err := store.DeleteExpiredRejectedApplications(ctx, ttlDays); err != nil {
			log.Printf("failed to cleanup expired rejected applications: %v", err)
		}

		for range ticker.C {
			if err := store.DeleteExpiredRejectedApplications(ctx, ttlDays); err != nil {
				log.Printf("failed to cleanup expired rejected applications: %v", err)
			}
		}
	}()

	addr := ":8080"
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
