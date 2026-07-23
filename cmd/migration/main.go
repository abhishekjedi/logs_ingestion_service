package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"error-logging/db/migration"
	mysqlclient "error-logging/pkg/client/mysql"
	"error-logging/pkg/config"

	"github.com/golang-migrate/migrate/v4"
	mysqldriver "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// The migration command applies all pending MySQL migrations, then exits. Run it
// as a separate step (locally or as a job) rather than on server startup:
//
//	go run ./cmd/migration
func main() {
	log.Println("Running MySQL migrations...")

	k := config.NewDefaultConfigProvider()
	client, err := mysqlclient.NewClient(config.NewMysqlConfig(k))
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("error closing mysql client: %v", err)
		}
	}()

	sqlDB, err := client.DB.DB()
	if err != nil {
		log.Fatalf("get sql db handle: %v", err)
	}

	if err := runMySQL(sqlDB); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Println("MySQL migrations up to date")
}

// runMySQL applies all pending MySQL migrations. Safe to run repeatedly:
// already-applied versions are skipped and ErrNoChange is treated as success.
func runMySQL(db *sql.DB) error {
	src, err := iofs.New(migration.MySQLFS, "mysql")
	if err != nil {
		return fmt.Errorf("load migration source: %w", err)
	}

	driver, err := mysqldriver.WithInstance(db, &mysqldriver.Config{})
	if err != nil {
		return fmt.Errorf("init mysql migrate driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply mysql migrations: %w", err)
	}
	return nil
}
