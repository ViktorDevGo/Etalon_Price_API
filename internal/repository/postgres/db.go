package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps pgxpool.Pool with additional functionality
type DB struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// Config holds database configuration
type Config struct {
	DSN            string
	MaxConnections int32
	MinConnections int32
	MaxConnLife    time.Duration
	MaxConnIdle    time.Duration
	Logger         *slog.Logger
}

// New creates a new database connection pool
func New(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	// Parse connection string
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database DSN: %w", err)
	}

	// Set pool configuration
	if cfg.MaxConnections > 0 {
		poolCfg.MaxConns = cfg.MaxConnections
	} else {
		poolCfg.MaxConns = 25
	}

	if cfg.MinConnections > 0 {
		poolCfg.MinConns = cfg.MinConnections
	} else {
		poolCfg.MinConns = 5
	}

	if cfg.MaxConnLife > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLife
	} else {
		poolCfg.MaxConnLifetime = 1 * time.Hour
	}

	if cfg.MaxConnIdle > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdle
	} else {
		poolCfg.MaxConnIdleTime = 30 * time.Minute
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Ping database
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	cfg.Logger.Info("Database connection established",
		"max_connections", poolCfg.MaxConns,
		"min_connections", poolCfg.MinConns,
	)

	return &DB{
		pool:   pool,
		logger: cfg.Logger,
	}, nil
}

// Pool returns the underlying connection pool
func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.pool.Close()
	db.logger.Info("Database connection closed")
}

// Ping checks database connectivity
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Stats returns pool statistics
func (db *DB) Stats() *pgxpool.Stat {
	return db.pool.Stat()
}
