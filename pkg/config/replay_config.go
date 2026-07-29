package config

import (
	"os"
	"time"

	"github.com/knadh/koanf/v2"
)

type ReplayConfig struct {
	EncryptionKey  string
	RequestTimeout time.Duration
	CacheTTL       time.Duration
	AllowInsecure  bool
}

func NewReplayConfig(k *koanf.Koanf) ReplayConfig {
	key := os.Getenv("ERRLOG_REPLAY_ENCRYPTION_KEY")
	if key == "" {
		key = k.String("replay.encryption_key")
	}
	return ReplayConfig{
		EncryptionKey:  key,
		RequestTimeout: kDuration(k, "replay.request_timeout", 5*time.Second),
		CacheTTL:       kDuration(k, "replay.cache_ttl", time.Minute),
		AllowInsecure:  k.Bool("replay.allow_insecure"),
	}
}
