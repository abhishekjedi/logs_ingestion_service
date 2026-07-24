package main

import (
	"context"
	"log"
	"math"
	"os/signal"
	"runtime/debug"
	"syscall"

	"error-logging/di"
	kafkaclient "error-logging/pkg/client/kafka"
	"error-logging/services"
)

func main() {
	log.Println("Starting Worker for error logging")

	// Surface the effective soft memory limit. Set GOMEMLIMIT (env, or the
	// container's) to make the GC defend a ceiling instead of risking an OOM.
	if limit := debug.SetMemoryLimit(-1); limit == math.MaxInt64 {
		log.Println("GOMEMLIMIT: unset (unlimited) — set GOMEMLIMIT to bound worker memory")
	} else {
		log.Printf("GOMEMLIMIT: %d bytes", limit)
	}

	container := di.BuildWorkerContainer()

	if err := container.Invoke(runWorker); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

// runWorker runs the batch consumer until a shutdown signal arrives, then closes
// Kafka. The consumer finishes any in-flight cycle on a detached context, so no
// buffered work is lost on a graceful stop.
func runWorker(consumer services.BatchConsumer, client *kafkaclient.Client) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("error closing kafka client: %v", err)
		}
	}()

	return consumer.Run(ctx)
}
