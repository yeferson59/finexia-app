// Package cache owns the Redis-backed fiber.Storage shared by every module.
// It is part of the platform layer and must stay free of business logic.
package cache

import (
	"github.com/gofiber/storage/redis/v3"
)

// Connect opens a Redis-backed fiber.Storage against url.
func Connect(url string) *redis.Storage {
	return redis.New(redis.Config{
		URL: url,
	})
}
