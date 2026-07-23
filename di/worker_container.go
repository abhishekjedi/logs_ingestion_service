package di

import (
	"error-logging/pkg/config"
	chclient "error-logging/pkg/client/clickhouse"
	kafkaclient "error-logging/pkg/client/kafka"
	mysqlclient "error-logging/pkg/client/mysql"
	redisclient "error-logging/pkg/client/redis"

	"go.uber.org/dig"
)

func BuildWorkerContainer() *dig.Container {
	container := dig.New()

	// Configs
	container.Provide(config.NewDefaultConfigProvider)
	container.Provide(config.NewMysqlConfig)
	container.Provide(config.NewRedisConfig)
	container.Provide(config.NewClickhouseConfig)
	container.Provide(config.NewKafkaConfig)

	// Clients
	container.Provide(mysqlclient.NewClient)
	container.Provide(redisclient.NewClient)
	container.Provide(chclient.NewClient)
	container.Provide(kafkaclient.NewClient)

	// Repos

	// Services

	return container
}
