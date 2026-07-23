package kafka

import (
	"error-logging/pkg/config"
	"log"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	Reader *kafka.Reader
	Writer *kafka.Writer
}

func NewClient(cfg config.KafkaConfig) *Client {
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
	return &Client{Reader: reader, Writer: writer}
}
