package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ziad-azm/ZCRM/backend/internal/config"
)

// Pinger reports whether a data store is reachable. Services depend on this
// interface rather than on *pgxpool.Pool so they stay testable without a
// database.
type Pinger interface {
	Ping(ctx context.Context) error
}

// NewPool returns a connection pool configured from cfg.
//
// It does NOT establish a connection: pgxpool connects lazily, so a database
// that is down produces no error here and surfaces on the first query or Ping.
// That is intentional — the API must still start and report an unhealthy
// database rather than refusing to boot.
func NewPool(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database dsn: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod
	poolCfg.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	return pool, nil
}
