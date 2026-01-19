package bootstrap

type StorageType string

const (
	StorageMemory   StorageType = "memory"
	StoragePostgres StorageType = "postgres"
)

type EmailService string

const (
	EmailServiceResend EmailService = "resend"
	EmailServiceNone   EmailService = "none"
)

type Config struct {
	Storage       StorageType
	EmailService  EmailService
	BCryptCost    int
	SeedUsers     bool
	SeedProposals bool
}
