package clickhouse

import (
	"fmt"
	"log"

	"error-logging/pkg/config"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type Client struct {
	DB *gorm.DB
}

func NewClient(cfg config.ClickhouseConfig) (*Client, error) {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?dial_timeout=10s&read_timeout=20s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	db, err := gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("connect clickhouse: %w", err)
	}

	log.Println("ClickHouse connected successfully")
	return &Client{DB: db}, nil
}

func (c *Client) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
