package main

import (
	"bytecourses/internal/database"
	"bytecourses/internal/handlers"
	"bytecourses/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    database.Connect()
    defer database.DB.Close()
    r := mux.NewRouter()
    authHandler := handlers.NewAuthHandler()
    
    r.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
    r.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
    
    api := r.PathPrefix("/api").Subrouter()
    api.Use(middleware.AuthMiddleware)
    api.HandleFunc("/auth/me", authHandler.GetCurrentUser).Methods("GET")

    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
    })

    handler := c.Handler(r)

    port := os.Getenv("PORT")
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}
