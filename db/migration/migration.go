// Package migration holds the versioned SQL schema migrations, embedded so they
// ship with the binary. This package represents the migrations themselves; the
// logic that applies them lives in cmd/migration.
package migration

import "embed"

// MySQLFS holds the MySQL migration files. It is co-located with the .sql files
// because go:embed paths are relative to the embedding source file.
//
//go:embed mysql/*.sql
var MySQLFS embed.FS

// ClickHouseFS holds the ClickHouse cluster migration files.
//
//go:embed clickhouse/*.sql
var ClickHouseFS embed.FS
