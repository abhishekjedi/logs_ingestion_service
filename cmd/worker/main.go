package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"error-logging/di"
	kafkaclient "error-logging/pkg/client/kafka"
)

func main() {
	log.Println("Starting Worker for error logging")

	container := di.BuildWorkerContainer()

	if err := container.Invoke(runWorker); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

// runWorker consumes Kafka messages until a shutdown signal arrives, then closes
// the Kafka client. The consume loop is driven by a cancellable context so an
// interrupt unblocks ReadMessage promptly instead of waiting on the next message.
func runWorker(client *kafkaclient.Client) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("error closing kafka client: %v", err)
		}
	}()

	log.Println("Worker started, consuming messages...")
	for {
		msg, err := client.Reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("shutdown signal received, stopping worker")
				return nil
			}
			log.Printf("error reading message: %v", err)
			continue
		}

		log.Printf("received message: topic=%s key=%s value=%s",
			msg.Topic, string(msg.Key), string(msg.Value),
		)
	}
}
