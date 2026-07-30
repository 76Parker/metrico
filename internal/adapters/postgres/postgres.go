package postgres

import (
	"github.com/76Parker/metrico/internal/adapters/postgres/pgen"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
	db   *pgen.Queries
	*healthRepo
	*metricsRepo
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:        pool,
		db:          pgen.New(pool),
		healthRepo:  &healthRepo{pool: pool},
		metricsRepo: &metricsRepo{q: pgen.New(pool)},
	}
}
