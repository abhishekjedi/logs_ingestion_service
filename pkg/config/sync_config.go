package config

import (
	"time"

	"github.com/knadh/koanf/v2"
)

type SyncConfig struct {
	Interval time.Duration

	ActiveWindow time.Duration

	BatchSize int
}

func NewSyncConfig(k *koanf.Koanf) SyncConfig {
	return SyncConfig{
		Interval:     envDuration("ERRLOG_SYNC_INTERVAL", kDuration(k, "sync.interval", 10*time.Second)),
		ActiveWindow: envDuration("ERRLOG_SYNC_ACTIVE_WINDOW", kDuration(k, "sync.active_window", 2*time.Hour)),
		BatchSize:    envInt("ERRLOG_SYNC_BATCH_SIZE", kInt(k, "sync.batch_size", 500)),
	}
}
