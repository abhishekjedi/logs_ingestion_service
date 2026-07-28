package config

import (
	"os"
	"strconv"
	"time"

	"github.com/knadh/koanf/v2"
)

type WorkerConfig struct {
	PoolSize int

	FetchMaxMessages int

	FetchMaxBytes int

	FetchMaxWait time.Duration

	FlushChunkRows int

	FlushRetries int
}

func NewWorkerConfig(k *koanf.Koanf) WorkerConfig {
	return WorkerConfig{
		PoolSize:         envInt("ERRLOG_WORKER_POOL_SIZE", kInt(k, "worker.pool_size", 16)),
		FetchMaxMessages: envInt("ERRLOG_WORKER_FETCH_MAX_MESSAGES", kInt(k, "worker.fetch_max_messages", 500)),
		FetchMaxBytes:    envInt("ERRLOG_WORKER_FETCH_MAX_BYTES", kInt(k, "worker.fetch_max_bytes", 32<<20)),
		FetchMaxWait:     envDuration("ERRLOG_WORKER_FETCH_MAX_WAIT", kDuration(k, "worker.fetch_max_wait", 200*time.Millisecond)),
		FlushChunkRows:   envInt("ERRLOG_WORKER_FLUSH_CHUNK_ROWS", kInt(k, "worker.flush_chunk_rows", 50000)),
		FlushRetries:     envInt("ERRLOG_WORKER_FLUSH_RETRIES", kInt(k, "worker.flush_retries", 3)),
	}
}

func kInt(k *koanf.Koanf, key string, def int) int {
	if k.Exists(key) {
		return k.Int(key)
	}
	return def
}

func kBool(k *koanf.Koanf, key string, def bool) bool {
	if k.Exists(key) {
		return k.Bool(key)
	}
	return def
}

func kDuration(k *koanf.Koanf, key string, def time.Duration) time.Duration {
	if k.Exists(key) {
		if d := k.Duration(key); d > 0 {
			return d
		}
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
