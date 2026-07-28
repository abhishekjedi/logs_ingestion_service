package main

import (
	"context"
	"io"
	"log"
	"os/signal"
	"syscall"
	"time"

	"error-logging/di"
	chclient "error-logging/pkg/client/clickhouse"
	mysqlclient "error-logging/pkg/client/mysql"
	"error-logging/pkg/config"
	"error-logging/services"
)

func main() {
	log.Println("Starting Syncer for error logging")

	container := di.BuildSyncContainer()

	if err := container.Invoke(runSyncer); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func runSyncer(
	sync services.SyncService,
	cfg config.SyncConfig,
	mysqlC *mysqlclient.Client,
	chC *chclient.NativeClient,
) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	defer closeAll(mysqlC, chC)

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	log.Printf("syncer running every %s (active window %s)", cfg.Interval, cfg.ActiveWindow)
	for {
		if n, err := sync.SyncOnce(ctx); err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("sync error: %v", err)
		} else if n > 0 {
			log.Printf("synced %d issues", n)
		}

		select {
		case <-ctx.Done():
			log.Println("syncer stopped cleanly")
			return nil
		case <-ticker.C:
		}
	}
	log.Println("syncer stopped cleanly")
	return nil
}

func closeAll(closers ...io.Closer) {
	for _, c := range closers {
		if c == nil {
			continue
		}
		if err := c.Close(); err != nil {
			log.Printf("error closing resource: %v", err)
		}
	}
}
