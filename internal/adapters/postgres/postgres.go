package postgres

import "github.com/jackc/pgx/v5"

type Repository struct {
	conn *pgx.Conn
	*healthRepository
}

func NewRepository(conn *pgx.Conn) *Repository {
	return &Repository{
		conn:             conn,
		healthRepository: &healthRepository{conn: conn},
	}
}
