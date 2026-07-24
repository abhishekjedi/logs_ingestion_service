package config

import (
	"os"
	"strconv"
	"time"

	"github.com/knadh/koanf/v2"
)

// WorkerConfig tunes the batch consumer. Values come from config.yaml, overridden
// by ERRLOG_WORKER_* environment variables (env wins), each with a safe default.
type WorkerConfig struct {
	// PoolSize is the transform parallelism within a cycle.
	PoolSize int
	// FetchMaxMessages caps how many Kafka messages a cycle accumulates.
	FetchMaxMessages int
	// FetchMaxBytes caps the total raw payload bytes a cycle accumulates, bounding
	// peak memory regardless of individual message size.
	FetchMaxBytes int
	// FetchMaxWait caps how long a cycle waits to fill (latency bound at low traffic).
	FetchMaxWait time.Duration
	// FlushChunkRows caps rows per ClickHouse insert so the driver's batch buffer
	// never holds an entire large cycle at once. 0 = no chunking.
	FlushChunkRows int
	// FlushRetries is how many times a failed ClickHouse flush chunk is retried.
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
