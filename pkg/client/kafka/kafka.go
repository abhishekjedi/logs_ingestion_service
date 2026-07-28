package kafka

import (
	"errors"
	"fmt"
	"log"
	"time"

	"error-logging/pkg/config"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	Reader *kafka.Reader
	Writer *kafka.Writer
}

func NewClient(cfg config.KafkaConfig) (*Client, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 1,
		MaxBytes: 10e6,

		SessionTimeout:    6 * time.Second,
		RebalanceTimeout:  6 * time.Second,
		HeartbeatInterval: 2 * time.Second,
	})

	writer := &kafka.Writer{
		Addr:  kafka.TCP(cfg.Brokers...),
		Topic: cfg.Topic,
		// Hash the message key (service id) so a service's records map to a stable
		// partition — per-service ordering plus even spread across partitions.
		Balancer: &kafka.Hash{},
		// Async: the ingest handler enqueues and returns 202 immediately instead of
		// blocking on the broker round-trip. This is what lets ingest sustain high
		// request rates; background failures surface via Completion. The writer is
		// flushed on shutdown (client.Close) so buffered records are not lost.
		Async:        true,
		BatchSize:    500,
		BatchTimeout: 50 * time.Millisecond,
		RequiredAcks: kafka.RequireOne,
		Completion: func(msgs []kafka.Message, err error) {
			if err != nil {
				log.Printf("kafka async write failed for %d message(s): %v", len(msgs), err)
			}
		},
	}

	log.Println("Kafka client initialized successfully")
	return &Client{Reader: reader, Writer: writer}, nil
}

func (c *Client) Close() error {
	var errs []error
	if c.Reader != nil {
		if err := c.Reader.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close reader: %w", err))
		}
	}
	if c.Writer != nil {
		if err := c.Writer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close writer: %w", err))
		}
	}
	return errors.Join(errs...)
}
