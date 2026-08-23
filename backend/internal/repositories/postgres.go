package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pinger reports whether a data store is reachable. Services depend on this
// interface rather than on *pgxpool.Pool so they stay testable without a
// database.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Pool configuration. These are deliberate defaults for a single API instance;
// ZCRM-6 makes them configurable.
const (
	maxConns          = 10
	minConns          = 2
	maxConnLifetime   = time.Hour
	maxConnIdleTime   = 30 * time.Minute
	healthCheckPeriod = time.Minute
	connectTimeout    = 5 * time.Second
)

// NewPool parses dsn and returns a configured connection pool.
//
// It does NOT establish a connection: pgxpool connects lazily, so a database
// that is down produces no error here and surfaces on the first query or Ping.
// That is intentional — the API must still start and report an unhealthy
// database rather than refusing to boot.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse database dsn: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns
	cfg.MaxConnLifetime = maxConnLifetime
	cfg.MaxConnIdleTime = maxConnIdleTime
	cfg.HealthCheckPeriod = healthCheckPeriod
	cfg.ConnConfig.ConnectTimeout = connectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	return pool, nil
}
