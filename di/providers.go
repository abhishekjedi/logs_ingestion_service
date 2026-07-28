package di

import (
	"log"

	redisclient "error-logging/pkg/client/redis"
	"error-logging/pkg/config"
)

func provideRedis(cfg config.RedisConfig) *redisclient.Client {
	client, err := redisclient.NewClient(cfg)
	if err != nil {
		log.Printf("WARNING: Redis unavailable, running with cache disabled until it recovers: %v", err)
	}
	return client
}
