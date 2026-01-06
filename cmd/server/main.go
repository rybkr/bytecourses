package main

import (
	"bytecourses/internal/app"
	"bytecourses/internal/auth"
	"context"
	"flag"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func main() {
	httpAddr := flag.String("http-addr", ":8080", "http listen address")
	storage := flag.String("storage", "memory", "storage backend: memory|sql")
	dbDsn := flag.String("database-dsn", "", "SQL database DSN (required if storage=sql)")
	bcryptCost := flag.Int("bcrypt-cost", bcrypt.DefaultCost, "bcrypt cost factor")
	seedUsers := flag.Bool("seed-users", false, "seed system test users")
	flag.Parse()

	ctx := context.Background()
	cfg := app.Config{
		HTTPAddr:    *httpAddr,
		Storage:     app.StorageBackend(*storage),
		DatabaseDSN: *dbDsn,
		BcryptCost:  *bcryptCost,
		SeedUsers:   *seedUsers,
	}
	auth.SetBcryptCost(*bcryptCost)

	a, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, a.Router()))
}
