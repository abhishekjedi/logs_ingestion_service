package clickhouse

import (
	"context"
	"fmt"
	"time"

	"error-logging/pkg/config"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// NativeClient is a clickhouse-go/v2 native connection used for high-throughput
// batch inserts (proper Map / Array(Tuple) handling), separate from the gorm client
// used for analytical reads.
type NativeClient struct {
	Conn driver.Conn
}

func NewNativeClient(cfg config.ClickhouseConfig) (*NativeClient, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.DBName,
			Username: cfg.User,
			Password: cfg.Password,
		},
		// Async inserts let ClickHouse coalesce many small concurrent inserts into
		// large parts server-side, avoiding the "too many parts" throttle under a
		// high-concurrency writer. wait_for_async_insert=1 keeps it durable: Send
		// returns once the row is flushed into a part.
		Settings: clickhouse.Settings{
			"async_insert":          1,
			"wait_for_async_insert": 1,
		},
		MaxOpenConns: 32,
		MaxIdleConns: 16,
	})
	if err != nil {
		return nil, fmt.Errorf("open clickhouse native: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping clickhouse native: %w", err)
	}

	return &NativeClient{Conn: conn}, nil
}

func (c *NativeClient) Close() error {
	return c.Conn.Close()
}
