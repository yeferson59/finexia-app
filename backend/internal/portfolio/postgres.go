package portfolio

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository implements Repository over the shared pgx pool. Its
// methods are split by sub-area across the postgres_*.go files.
type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return new(PostgresRepository{db})
}
