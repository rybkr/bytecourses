package app

type StorageBackend string

const (
	StorageMemroy StorageBackend = "memory"
	StorageSQL    StorageBackend = "sql"
)

type EmailServiceProvider string

const (
	EmailServiceNone   EmailServiceProvider = "none"
	EmailServiceResend EmailServiceProvider = "resend"
)

type Config struct {
	Storage      StorageBackend
	DatabaseDSN  string
	BcryptCost   int
	SeedUsers    bool
	EmailService EmailServiceProvider
}
