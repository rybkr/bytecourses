package store

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type Store struct {
	db *pgxpool.Pool
}

func New(connString string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Printf("failed to create connection pool: %v", err)
		return nil, err
	}

	if err := pool.Ping(context.Background()); err != nil {
		log.Printf("failed to ping database: %v", err)
		return nil, err
	}

	log.Println("database connection established")
	return &Store{db: pool}, nil
}

func (s *Store) Close() {
	s.db.Close()
	log.Println("database connection closed")
}
