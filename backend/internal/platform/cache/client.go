package cache

import "github.com/redis/go-redis/v9"

func Conn(host, port, password string, db int) *redis.Client {
	return redis.NewClient(new(redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       db,
	}))
}
