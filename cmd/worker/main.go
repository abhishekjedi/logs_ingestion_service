package main

import (
	"context"
	"error-logging/di"
	kafkaclient "error-logging/pkg/client/kafka"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Starting Worker for error logging")

	container := di.BuildWorkerContainer()

	if err := container.Invoke(startWorker); err != nil {
		log.Fatal(err)
	}
}

func startWorker(client *kafkaclient.Client) {
	ctx := context.Background()

	log.Println("Worker started, consuming messages...")
	for {
		msg, err := client.Reader.ReadMessage(ctx)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			continue
		}

		log.Printf("Received message: topic=%s key=%s value=%s",
			msg.Topic, string(msg.Key), string(msg.Value),
		)
	}
}
