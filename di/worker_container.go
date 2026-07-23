package di

import (
	chclient "error-logging/pkg/client/clickhouse"
	kafkaclient "error-logging/pkg/client/kafka"
	mysqlclient "error-logging/pkg/client/mysql"
	s3client "error-logging/pkg/client/s3"
	"error-logging/pkg/config"

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
	container.Provide(config.NewS3Config)

	// Clients (mysql/clickhouse/kafka required — errors fail startup; redis degradable)
	container.Provide(mysqlclient.NewClient)
	container.Provide(provideRedis)
	container.Provide(chclient.NewClient)
	container.Provide(kafkaclient.NewClient)
	container.Provide(s3client.NewClient)

	// Repos

	// Services

	return container
}
