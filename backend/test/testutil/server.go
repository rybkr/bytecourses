package testutil

import (
	"bytecourses/internal/database"
	"bytecourses/internal/handlers"
	"bytecourses/internal/middleware"
	"database/sql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"net/http/httptest"
	"os"
	"testing"
)

func SetupTestServer(t *testing.T) *httptest.Server {
    setupTestDB(t)

    r := mux.NewRouter()
    authHandler := handlers.NewAuthHandler()

    r.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
    r.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

    api := r.PathPrefix("/api").Subrouter()
    api.Use(middleware.AuthMiddleware)
    api.HandleFunc("/auth/me", authHandler.GetCurrentUser).Methods("GET")

    ts := httptest.NewServer(r)
    t.Cleanup(func() {
        ts.Close()
    })
    return ts
}

func setupTestDB(t *testing.T) {
    dbURL := os.Getenv("TEST_DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://ryanbaker@localhost:5432/bytecourses_test?sslmode=disable"
    }

    var err error
    database.DB, err = sql.Open("postgres", dbURL)
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }

    if err = database.DB.Ping(); err != nil {
        t.Fatalf("Failed to ping test database: %v", err)
    }

    CleanupTestData(t)
}

func CleanupTestData(t *testing.T) {
    queries := []string{
        "DELETE FROM users",
    }

    for _, query := range queries {
        if _, err := database.DB.Exec(query); err != nil {
            t.Logf("Warning: cleanup query failed: %v", err)
        }
    }
}

func CreateTestUser(t *testing.T, serverURL, email, password, name string) string {
    client := NewTestClient()
    
    body := map[string]string{
        "email":    email,
        "password": password,
        "name":     name,
    }

    var response struct {
        Token string `json:"token"`
    }
    err := client.Post(serverURL+"/api/auth/register", body, &response)
    if err != nil {
        t.Fatalf("Failed to create test user: %v", err)
    }

    return response.Token
}
