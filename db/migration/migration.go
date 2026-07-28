package migration

import "embed"

//go:embed mysql/*.sql
var MySQLFS embed.FS

//go:embed clickhouse/*.sql
var ClickHouseFS embed.FS
