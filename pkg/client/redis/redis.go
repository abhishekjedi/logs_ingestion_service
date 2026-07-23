package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"error-logging/pkg/config"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	RDB *redis.Client
}

// NewClient builds a Redis client and probes it. Redis is used only as a cache, so
// it is a degradable dependency: on a failed probe the returned *Client is still
// usable (go-redis reconnects lazily) and the error is returned for the caller to
// decide whether to treat the cache as temporarily unavailable rather than fatal.
func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	client := &Client{RDB: rdb}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return client, fmt.Errorf("ping redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return client, nil
}

// Close releases the client's connection pool.
func (c *Client) Close() error {
	return c.RDB.Close()
}
