package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
    db *sql.DB
}

type Config struct {
    DSN string
}

func NewStore(cfg Config) (*Store, error) {
    db, err := sql.Open("pgx", cfg.DSN)
    if err != nil {
        return nil, fmt.Errorf("open postgres: %w", err)
    }

    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxIdleTime(5 *time.Minute)
    db.SetConnMaxLifetime(60 * time.Minute)

    ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()

    err = db.PingContext(ctx)
     if err != nil {
         db.Close()
         return nil, fmt.Errorf("ping postgres: %w", err)
     }

    s := &Store{db: db}

    err = s.initSchema(ctx)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("init schema: %w", err)
    }

    return s, nil
}

func (s *Store) initSchema(ctx context.Context) error {
    const query = `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL UNIQUE,
        email TEXT NOT NULL,
        password_hash TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );`

    ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
    defer cancel()

    _, err := s.db.ExecContext(ctx, query)
    if err != nil {
        return fmt.Errorf("create users table: %w", err)
    }

    return nil
}

func (s *Store) Close() error {
    return s.db.Close()
}