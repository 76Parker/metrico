package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type healthRepository struct {
	conn *pgx.Conn
}

func (r *healthRepository) Ping(ctx context.Context) error {
	return r.conn.Ping(ctx)
}
