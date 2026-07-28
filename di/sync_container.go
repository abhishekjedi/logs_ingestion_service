package di

import (
	repositoryimpl "error-logging/db/repository/impl"
	chclient "error-logging/pkg/client/clickhouse"
	mysqlclient "error-logging/pkg/client/mysql"
	"error-logging/pkg/config"
	serviceimpl "error-logging/services/impl"

	"go.uber.org/dig"
)

func BuildSyncContainer() *dig.Container {
	container := dig.New()

	container.Provide(config.NewDefaultConfigProvider)
	container.Provide(config.NewMysqlConfig)
	container.Provide(config.NewClickhouseConfig)
	container.Provide(config.NewSyncConfig)

	container.Provide(mysqlclient.NewClient)
	container.Provide(chclient.NewNativeClient)

	container.Provide(repositoryimpl.NewIssueRepository)
	container.Provide(repositoryimpl.NewIssueStatsRepository)

	container.Provide(serviceimpl.NewSyncService)

	return container
}
