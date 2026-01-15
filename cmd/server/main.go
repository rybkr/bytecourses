package main

import (
	"bytecourses/internal/app"
	"bytecourses/internal/auth"
	"context"
	"flag"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
)

func main() {
	storage := flag.String("storage", "memory", "storage backend: memory|sql")
	bcryptCost := flag.Int("bcrypt-cost", bcrypt.DefaultCost, "bcrypt cost factor")
	emailService := flag.String("email-service", "none", "email service provider: resend|none")
	seedUsers := flag.Bool("seed-users", false, "seed system test users")
	seedProposals := flag.Bool("seed-proposals", false, "seed system test proposals")
	flag.Parse()

	ctx := context.Background()
	cfg := app.Config{
		Storage:       app.StorageBackend(*storage),
		BcryptCost:    *bcryptCost,
		SeedUsers:     *seedUsers,
		SeedProposals: *seedProposals,
		EmailService:  app.EmailServiceProvider(*emailService),
	}
	auth.SetBcryptCost(*bcryptCost)

	a, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := "0.0.0.0:" + port
	log.Fatal(http.ListenAndServe(addr, a.Router()))
}
