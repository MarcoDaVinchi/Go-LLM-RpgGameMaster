package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	HealthCheckPeriod time.Duration
	AcquireTimeout    time.Duration
}

func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   30 * time.Minute,
		HealthCheckPeriod: 5 * time.Minute,
		AcquireTimeout:    5 * time.Second,
	}
}

// NewPool creates a new connection pool with configuration
func NewPool(ctx context.Context, connString string, config *PoolConfig) (*pgxpool.Pool, error) {
	if config == nil {
		config = DefaultPoolConfig()
	}

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parsing connection string: %w", err)
	}

	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, config.AcquireTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	log.Info().
		Int32("max_conns", config.MaxConns).
		Int32("min_conns", config.MinConns).
		Msg("PostgreSQL connection pool created")

	return pool, nil
}

// HealthCheck checks if the database is healthy
func (r *PostgresRetriever) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := r.db.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	stats := r.db.Stat()
	log.Debug().
		Int32("total_conns", stats.TotalConns()).
		Int32("idle_conns", stats.IdleConns()).
		Int32("acquired_conns", stats.AcquiredConns()).
		Msg("Pool health check")
	return nil
}
