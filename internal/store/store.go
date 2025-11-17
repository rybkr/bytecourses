package store

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func New(connString string) (*Store, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	return &Store{db: pool}, nil
}

func (s *Store) Close() {
	s.db.Close()
}
