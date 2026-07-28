package main

import (
	"context"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"error-logging/di"
	kafkaclient "error-logging/pkg/client/kafka"
	"error-logging/services"
)

func main() {
	log.Println("Starting Worker for error logging")

	if addr := os.Getenv("PPROF_ADDR"); addr != "" {
		go func() {
			log.Printf("pprof listening on %s", addr)
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Printf("pprof server stopped: %v", err)
			}
		}()
	}

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
