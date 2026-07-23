package config

import "github.com/knadh/koanf/v2"

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func NewRedisConfig(k *koanf.Koanf) RedisConfig {
	return RedisConfig{
		Host:     k.String("redis.host"),
		Port:     k.String("redis.port"),
		Password: k.String("redis.password"),
		DB:       k.Int("redis.db"),
	}
}
