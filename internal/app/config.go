package app

type StorageBackend string

const (
	StorageMemroy StorageBackend = "memory"
	StorageSQL    StorageBackend = "sql"
)

type Config struct {
	HTTPAddr    string
	Storage     StorageBackend
	DatabaseDSN string
	BcryptCost  int
	SeedUsers   bool
}
