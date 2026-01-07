package sqlstore

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &Store{
		db: db,
	}, nil
}

func (s *Store) Close() error {
    return s.db.Close()
}
