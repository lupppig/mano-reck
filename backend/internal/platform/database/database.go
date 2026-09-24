// Package database owns the PostgreSQL pool and transaction boundary used by
// application use cases and repositories.
package database

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Config defines the bounded PostgreSQL pool policy.
type Config struct {
	ConnectionString string
	MaxConnections   int32
	MinConnections   int32
}

// Database owns one process-wide PostgreSQL connection pool.
type Database struct {
	config Config

	mu   sync.RWMutex
	pool *pgxpool.Pool
}

// New validates the pool configuration without opening a network connection.
func New(configuration Config) (*Database, error) {
	if configuration.MaxConnections < 1 {
		return nil, fmt.Errorf("database maximum connections must be positive")
	}
	if configuration.MinConnections < 0 || configuration.MinConnections > configuration.MaxConnections {
		return nil, fmt.Errorf("database minimum connections must be between zero and the maximum")
	}
	if _, err := pgxpool.ParseConfig(configuration.ConnectionString); err != nil {
		return nil, errors.New("parse database connection configuration")
	}
	return &Database{config: configuration}, nil
}

// Name identifies this dependency in readiness responses.
func (*Database) Name() string {
	return "postgres"
}

// Start creates the lazy pool. Connectivity is reported by Check so an
// ordinary database outage fails readiness without terminating the process.
func (database *Database) Start(ctx context.Context) error {
	database.mu.Lock()
	defer database.mu.Unlock()

	if database.pool != nil {
		return fmt.Errorf("database pool has already started")
	}

	poolConfig, err := pgxpool.ParseConfig(database.config.ConnectionString)
	if err != nil {
		return errors.New("parse database connection configuration")
	}
	poolConfig.MaxConns = database.config.MaxConnections
	poolConfig.MinConns = database.config.MinConnections

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	database.pool = pool
	return nil
}

// Check verifies that PostgreSQL can answer a bounded readiness query.
func (database *Database) Check(ctx context.Context) error {
	pool, err := database.Pool()
	if err != nil {
		return err
	}
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}

// Close releases every pooled connection.
func (database *Database) Close(context.Context) error {
	database.mu.Lock()
	pool := database.pool
	database.pool = nil
	database.mu.Unlock()

	if pool != nil {
		pool.Close()
	}
	return nil
}

// Pool returns the started pool for repository construction.
func (database *Database) Pool() (*pgxpool.Pool, error) {
	database.mu.RLock()
	defer database.mu.RUnlock()
	if database.pool == nil {
		return nil, fmt.Errorf("database pool is not started")
	}
	return database.pool, nil
}

// DBTX is the query surface shared by pgxpool.Pool and pgx.Tx. Repositories
// accept this boundary so application services can provide their transaction.
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}
