package config

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/storage/redis/v3"
)

func (Config) ConnectionCache(url string) fiber.Storage {
	return redis.New(redis.Config{
		URL: url,
	})
}
