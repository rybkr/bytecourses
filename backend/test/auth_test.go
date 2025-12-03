package test

import (
	"bytecourses/test/testutil"
	"github.com/joho/godotenv"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
    godotenv.Load("../.env")
    os.Setenv("JWT_SECRET", "test-secret-key")
    code := m.Run()
    os.Exit(code)
}

func TestRegisterUser(t *testing.T) {
    ts := testutil.SetupTestServer(t)
    client := testutil.NewTestClient()

    t.Run("successful registration", func(t *testing.T) {
        body := map[string]string{
            "email":    "user@example.com",
            "password": "password123",
            "name":     "New User",
        }

        var response struct {
            Token string `json:"token"`
            User  struct {
                Email string `json:"email"`
                Role  string `json:"role"`
            } `json:"user"`
        }

        err := client.Post(ts.URL+"/api/auth/register", body, &response)
        if err != nil {
            t.Fatalf("Registration failed: %v", err)
        }

        if response.Token == "" {
            t.Error("Expected token, got empty string")
        }

        if response.User.Email != "user@example.com" {
            t.Errorf("Expected email user@example.com, got %s", response.User.Email)
        }

        if response.User.Role != "student" {
            t.Errorf("Expected role student, got %s", response.User.Role)
        }
    })

    t.Run("duplicate email", func(t *testing.T) {
        body := map[string]string{
            "email":    "duplicate@example.com",
            "password": "password123",
            "name":     "First User",
        }

        var response struct {
            Token string `json:"token"`
        }

        err := client.Post(ts.URL+"/api/auth/register", body, &response)
        if err != nil {
            t.Fatalf("First registration failed: %v", err)
        }

        err = client.Post(ts.URL+"/api/auth/register", body, &response)
        if err == nil {
            t.Error("Expected error for duplicate email, got nil")
        }
    })
}

func TestLogin(t *testing.T) {
    ts := testutil.SetupTestServer(t)
    client := testutil.NewTestClient()

    email := "user@example.com"
    password := "password123"
    testutil.CreateTestUser(t, ts.URL, email, password, "Login Test")

    t.Run("successful login", func(t *testing.T) {
        body := map[string]string{
            "email":    email,
            "password": password,
        }

        var response struct {
            Token string `json:"token"`
            User  struct {
                Email string `json:"email"`
            } `json:"user"`
        }

        err := client.Post(ts.URL+"/api/auth/login", body, &response)
        if err != nil {
            t.Fatalf("Login failed: %v", err)
        }

        if response.Token == "" {
            t.Error("Expected token, got empty string")
        }

        if response.User.Email != email {
            t.Errorf("Expected email %s, got %s", email, response.User.Email)
        }
    })

    t.Run("wrong password", func(t *testing.T) {
        body := map[string]string{
            "email":    email,
            "password": "wrongpassword",
        }

        var response struct {
            Token string `json:"token"`
        }

        err := client.Post(ts.URL+"/api/auth/login", body, &response)
        if err == nil {
            t.Error("Expected error for wrong password, got nil")
        }
    })

    t.Run("nonexistent user", func(t *testing.T) {
        body := map[string]string{
            "email":    "nonexistent@purdue.edu",
            "password": "password123",
        }

        var response struct {
            Token string `json:"token"`
        }

        err := client.Post(ts.URL+"/api/auth/login", body, &response)
        if err == nil {
            t.Error("Expected error for nonexistent user, got nil")
        }
    })
}

func TestGetCurrentUser(t *testing.T) {
    ts := testutil.SetupTestServer(t)
    
    email := "currentuser@example.com"
    token := testutil.CreateTestUser(t, ts.URL, email, "password123", "Current User")

    t.Run("with valid token", func(t *testing.T) {
        client := testutil.NewTestClient().WithToken(token)

        var user struct {
            Email string `json:"email"`
            Name  string `json:"name"`
            Role  string `json:"role"`
        }

        err := client.Get(ts.URL+"/api/auth/me", &user)
        if err != nil {
            t.Fatalf("Failed to get current user: %v", err)
        }

        if user.Email != email {
            t.Errorf("Expected email %s, got %s", email, user.Email)
        }

        if user.Role != "student" {
            t.Errorf("Expected role student, got %s", user.Role)
        }
    })

    t.Run("without token", func(t *testing.T) {
        client := testutil.NewTestClient()

        var user struct {
            Email string `json:"email"`
        }

        err := client.Get(ts.URL+"/api/auth/me", &user)
        if err == nil {
            t.Error("Expected error without token, got nil")
        }
    })

    t.Run("with invalid token", func(t *testing.T) {
        client := testutil.NewTestClient().WithToken("invalid.token.here")

        var user struct {
            Email string `json:"email"`
        }

        err := client.Get(ts.URL+"/api/auth/me", &user)
        if err == nil {
            t.Error("Expected error with invalid token, got nil")
        }
    })
}
