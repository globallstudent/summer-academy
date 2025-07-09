package database

import (
	"context"
	"fmt"
	"time"

	"github.com/globallstudent/academy/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents a database connection pool
type DB struct {
	Pool *pgxpool.Pool
}

// Connect creates a new database connection pool
func Connect(cfg config.DatabaseConfig) (*DB, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
