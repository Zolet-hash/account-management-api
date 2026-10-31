package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Open function is responsible for creating and configuring Postgres conn pool
func Open(ctx context.Context, url string, tracer pgx.QueryTracer) (*pgxpool.Pool, error) {

	// Parse Database configuration
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}

	cfg.ConnConfig.Tracer = tracer //monitor database queries

	// configure conn pool
	cfg.MaxConns = 20                               //Limit maximum conn pool
	cfg.MinConns = 2                                //maintain atleast 2 conns
	cfg.MaxConnLifetime = 30 * time.Minute          //a connection cannot live longer than 30 sec
	cfg.MaxConnIdleTime = 5 * time.Minute           //Limit max iddle conn period
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second //Timeout when POSTGRESQL is unavailable or unreachable

	return pgxpool.NewWithConfig(ctx, cfg)

}
