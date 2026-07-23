package kafka

import (
	"errors"
	"fmt"
	"log"

	"error-logging/pkg/config"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	Reader *kafka.Reader
	Writer *kafka.Writer
}

// NewClient builds the Kafka reader and writer. kafka-go connects lazily on first
// use, so construction itself does not reach the broker.
func NewClient(cfg config.KafkaConfig) (*Client, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		GroupID: cfg.GroupID,
		Topic:   cfg.Topic,
	})

	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	log.Println("Kafka client initialized successfully")
	return &Client{Reader: reader, Writer: writer}, nil
}

// Close shuts down the reader and writer, joining any errors.
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
