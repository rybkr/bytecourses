package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"

	"bytecourses/internal/bootstrap"
	infraauth "bytecourses/internal/infrastructure/auth"
	infrahttp "bytecourses/internal/infrastructure/http"
	"bytecourses/web"
)

func main() {
	storage := flag.String("storage", "memory", "storage backend: memory|sql")
	bcryptCost := flag.Int("bcrypt-cost", bcrypt.DefaultCost, "bcrypt cost factor")
	emailService := flag.String("email-service", "none", "email service provider: resend|none")
	seedUsers := flag.String("seed-users", "", "path to JSON file containing users to seed")
	seedProposals := flag.String("seed-proposals", "", "path to JSON file containing proposals to seed")
	seedCourses := flag.String("seed-courses", "", "path to JSON file containing courses to seed")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	infraauth.SetBcryptCost(*bcryptCost)

	var storageType bootstrap.StorageType
	switch *storage {
	case "memory":
		storageType = bootstrap.StorageMemory
	case "sql", "postgres":
		storageType = bootstrap.StoragePostgres
	default:
		logger.Error("unknown storage type", "storage", *storage)
		os.Exit(1)
	}

	var emailServiceType bootstrap.EmailService
	switch *emailService {
	case "resend":
		emailServiceType = bootstrap.EmailServiceResend
	case "none":
		emailServiceType = bootstrap.EmailServiceNone
	default:
		logger.Error("unknown email service", "email-service", *emailService)
		os.Exit(1)
	}

	cfg := bootstrap.Config{
		Storage:       storageType,
		EmailService:  emailServiceType,
		BCryptCost:    *bcryptCost,
		SeedUsers:     *seedUsers,
		SeedProposals: *seedProposals,
		SeedCourses:   *seedCourses,
	}

	ctx := context.Background()

	container, err := bootstrap.NewContainer(ctx, cfg)
	if err != nil {
		logger.Error("failed to create container", "error", err)
		os.Exit(1)
	}
	defer container.Close()

	router := infrahttp.NewRouter(container, web.FS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := infrahttp.NewServer(router, port, logger)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			logger.Error("server error", "error", err)
		}
	}()

	logger.Info("server started", "port", port, "storage", storageType)

	<-done
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
