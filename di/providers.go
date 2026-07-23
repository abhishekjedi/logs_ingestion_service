package di

import (
	"log"

	redisclient "error-logging/pkg/client/redis"
	"error-logging/pkg/config"
)

// provideRedis wires Redis as a degradable dependency. Redis backs only caching,
// so a connection failure must not stop the process from booting: we log a warning
// and hand back the (still usable) client, which reconnects lazily once Redis
// recovers. Required dependencies (MySQL, ClickHouse, Kafka) are provided via their
// constructors directly, so their errors propagate and fail startup.
func provideRedis(cfg config.RedisConfig) *redisclient.Client {
	client, err := redisclient.NewClient(cfg)
	if err != nil {
		log.Printf("WARNING: Redis unavailable, running with cache disabled until it recovers: %v", err)
	}
	return client
}
