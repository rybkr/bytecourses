package main

import (
	"github.com/rybkr/bytecourses/internal/handlers"
	"github.com/rybkr/bytecourses/internal/store"
	"log"
	"net/http"
	"os"
)

func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		connString = "postgres://localhost/byte_course?sslmode=disable"
	}

	store, err := store.New(connString)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	courseHandler := handlers.NewCourseHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/courses", courseHandler.CreateCourse)
	mux.HandleFunc("GET /api/courses", courseHandler.ListCourses)
	mux.HandleFunc("PATCH /api/courses/approve", courseHandler.ApproveCourse)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
