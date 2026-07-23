package config

import "github.com/knadh/koanf/v2"

type KafkaConfig struct {
	Brokers []string
	GroupID string
	Topic   string
}

func NewKafkaConfig(k *koanf.Koanf) KafkaConfig {
	return KafkaConfig{
		Brokers: k.Strings("kafka.brokers"),
		GroupID: k.String("kafka.group_id"),
		Topic:   k.String("kafka.topic"),
	}
}
