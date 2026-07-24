package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type healthRepo struct {
	pool *pgxpool.Pool
}

func (r *healthRepo) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}
