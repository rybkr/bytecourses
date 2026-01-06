package main

import (
	"bytecourses/internal/app"
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
	"flag"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func ensureTestAdmin(ctx context.Context, users store.UserStore) error {
	email := "admin@local.bytecourses.org"
	if _, ok := users.GetUserByEmail(ctx, email); ok {
		return nil
	}

	hash, err := auth.HashPassword("admin")
	if err != nil {
		return err
	}

	return users.InsertUser(ctx, &domain.User{
		Email:        email,
		PasswordHash: hash,
		Role:         domain.UserRoleAdmin,
		Name:         "Admin User",
	})
}

func main() {
	httpAddr := flag.String("http-addr", ":8080", "http listen address")
	storage := flag.String("storage", "memory", "storage backend: memory|sql")
	dbDsn := flag.String("database-dsn", "", "SQL database DSN (required if storage=sql)")
	bcryptCost := flag.Int("bcrypt-cost", bcrypt.DefaultCost, "bcrypt cost factor")
	seedAdmin := flag.Bool("seed-admin", false, "seed a test admin user")
	flag.Parse()

	ctx := context.Background()
	cfg := app.Config{
		HTTPAddr:    *httpAddr,
		Storage:     app.StorageBackend(*storage),
		DatabaseDSN: *dbDsn,
		BcryptCost:  *bcryptCost,
		SeedAdmin:   *seedAdmin,
	}
	auth.SetBcryptCost(*bcryptCost)

	a, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.SeedAdmin {
		if err := ensureTestAdmin(ctx, a.UserStore); err != nil {
			log.Fatal(err)
		}
	}

    log.Printf("listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, a.Router()))
}
