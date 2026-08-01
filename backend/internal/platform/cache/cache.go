// Package cache owns the Redis-backed fiber.Storage shared by every module.
// It is part of the platform layer and must stay free of business logic.
package cache

import (
	"context"

	fibeRedis "github.com/gofiber/storage/redis/v3"
	"github.com/redis/go-redis/v9"
)

// Connect opens a Redis-backed fiber.Storage against url.
func Storage(ctx context.Context, conn redis.UniversalClient) *fibeRedis.Storage {
	return fibeRedis.NewFromConnection(conn)
}
