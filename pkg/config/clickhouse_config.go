package config

import "github.com/knadh/koanf/v2"

type ClickhouseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func NewClickhouseConfig(k *koanf.Koanf) ClickhouseConfig {
	return ClickhouseConfig{
		Host:     k.String("clickhouse.host"),
		Port:     k.String("clickhouse.port"),
		User:     k.String("clickhouse.user"),
		Password: k.String("clickhouse.password"),
		DBName:   k.String("clickhouse.dbname"),
	}
}
