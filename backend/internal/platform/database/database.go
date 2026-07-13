// Package database owns the PostgreSQL connection pool and transaction
// helpers shared by every module. It is part of the platform layer and must
// stay free of business logic.
package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pgx connection pool against databaseURL.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, databaseURL)
}
