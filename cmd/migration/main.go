package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"error-logging/db/migration"
	mysqlclient "error-logging/pkg/client/mysql"
	"error-logging/pkg/config"

	_ "github.com/ClickHouse/clickhouse-go/v2" // registers the "clickhouse" sql driver
	"github.com/golang-migrate/migrate/v4"
	chdriver "github.com/golang-migrate/migrate/v4/database/clickhouse"
	mysqldriver "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// The migration command applies all pending migrations across every store, then
// exits. Run it as a separate step (locally or as a job), not on server startup:
//
//	go run ./cmd/migration
func main() {
	log.Println("Running migrations...")

	k := config.NewDefaultConfigProvider()

	if err := runMySQL(config.NewMysqlConfig(k)); err != nil {
		log.Fatalf("mysql migrations: %v", err)
	}
	if err := runClickHouse(config.NewClickhouseConfig(k)); err != nil {
		log.Fatalf("clickhouse migrations: %v", err)
	}

	log.Println("All migrations up to date")
}

func runMySQL(cfg config.MysqlConfig) error {
	client, err := mysqlclient.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("connect mysql: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("error closing mysql client: %v", err)
		}
	}()

	sqlDB, err := client.DB.DB()
	if err != nil {
		return fmt.Errorf("obtain mysql handle: %w", err)
	}

	src, err := iofs.New(migration.MySQLFS, "mysql")
	if err != nil {
		return fmt.Errorf("load mysql migration source: %w", err)
	}

	driver, err := mysqldriver.WithInstance(sqlDB, &mysqldriver.Config{})
	if err != nil {
		return fmt.Errorf("init mysql migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("init mysql migrate: %w", err)
	}
	return applyUp(m)
}

func runClickHouse(cfg config.ClickhouseConfig) error {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	conn, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return fmt.Errorf("open clickhouse: %w", err)
	}
	defer conn.Close()

	src, err := iofs.New(migration.ClickHouseFS, "clickhouse")
	if err != nil {
		return fmt.Errorf("load clickhouse migration source: %w", err)
	}

	// MultiStatementEnabled: each migration file bundles several ON CLUSTER DDL
	// statements. ClusterName is left empty so migrate's own bookkeeping table
	// stays local — schema DDL propagates across nodes via the ON CLUSTER clauses.
	driver, err := chdriver.WithInstance(conn, &chdriver.Config{
		DatabaseName:          cfg.DBName,
		MultiStatementEnabled: true,
	})
	if err != nil {
		return fmt.Errorf("init clickhouse migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "clickhouse", driver)
	if err != nil {
		return fmt.Errorf("init clickhouse migrate: %w", err)
	}
	return applyUp(m)
}

// applyUp runs all pending migrations, treating "no change" as success.
func applyUp(m *migrate.Migrate) error {
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
