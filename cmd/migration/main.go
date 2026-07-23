package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"error-logging/db/migration"
	"error-logging/pkg/config"

	_ "github.com/ClickHouse/clickhouse-go/v2" // registers the "clickhouse" sql driver
	_ "github.com/go-sql-driver/mysql"         // registers the "mysql" sql driver
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
	// A dedicated connection with multiStatements=true so a migration file may
	// contain several ;-separated statements. This is kept separate from the app's
	// gorm pool, which deliberately does NOT enable stacked statements.
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true&parseTime=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open mysql: %w", err)
	}
	defer conn.Close()

	src, err := iofs.New(migration.MySQLFS, "mysql")
	if err != nil {
		return fmt.Errorf("load mysql migration source: %w", err)
	}

	driver, err := mysqldriver.WithInstance(conn, &mysqldriver.Config{})
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
