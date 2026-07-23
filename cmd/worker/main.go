package main

import (
	"context"
	"encoding/json"
	"log"
	"os/signal"
	"syscall"

	"error-logging/di"
	"error-logging/dto"
	kafkaclient "error-logging/pkg/client/kafka"
	"error-logging/services"
)

func main() {
	log.Println("Starting Worker for error logging")

	container := di.BuildWorkerContainer()

	if err := container.Invoke(runWorker); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

// runWorker consumes ingest messages and processes each through the pipeline until
// a shutdown signal arrives, then closes the Kafka client.
func runWorker(client *kafkaclient.Client, processor services.ProcessorService) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("error closing kafka client: %v", err)
		}
	}()

	log.Println("Worker started, consuming messages...")
	for {
		m, err := client.Reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("shutdown signal received, stopping worker")
				return nil
			}
			log.Printf("error reading message: %v", err)
			continue
		}

		var msg dto.LogIngestMessage
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			log.Printf("skipping malformed message: %v", err)
			continue
		}

		if err := processor.Process(ctx, msg); err != nil {
			log.Printf("process message (service=%d): %v", msg.ServiceID, err)
			continue
		}
		log.Printf("processed batch service=%d project=%d", msg.ServiceID, msg.ProjectID)
	}
}
