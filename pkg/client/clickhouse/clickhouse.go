package clickhouse

import (
	"error-logging/pkg/config"
	"fmt"
	"log"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type Client struct {
	DB *gorm.DB
}

func NewClient(cfg config.ClickhouseConfig) *Client {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?dial_timeout=10s&read_timeout=20s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to ClickHouse: %v", err)
	}

	log.Println("ClickHouse connected successfully")
	return &Client{DB: db}
}
