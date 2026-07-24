package di

import (
	repositoryimpl "error-logging/db/repository/impl"
	chclient "error-logging/pkg/client/clickhouse"
	mysqlclient "error-logging/pkg/client/mysql"
	"error-logging/pkg/config"
	serviceimpl "error-logging/services/impl"

	"go.uber.org/dig"
)

// BuildSyncContainer wires the standalone syncer: it needs MySQL (write) and the
// ClickHouse native client (read) only.
func BuildSyncContainer() *dig.Container {
	container := dig.New()

	// Configs
	container.Provide(config.NewDefaultConfigProvider)
	container.Provide(config.NewMysqlConfig)
	container.Provide(config.NewClickhouseConfig)
	container.Provide(config.NewSyncConfig)

	// Clients
	container.Provide(mysqlclient.NewClient)
	container.Provide(chclient.NewNativeClient)

	// Repositories
	container.Provide(repositoryimpl.NewIssueRepository)
	container.Provide(repositoryimpl.NewIssueStatsRepository)

	// Services
	container.Provide(serviceimpl.NewSyncService)

	return container
}
