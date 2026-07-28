package di

import (
	repositoryimpl "error-logging/db/repository/impl"
	chclient "error-logging/pkg/client/clickhouse"
	kafkaclient "error-logging/pkg/client/kafka"
	mysqlclient "error-logging/pkg/client/mysql"
	redisclient "error-logging/pkg/client/redis"
	s3client "error-logging/pkg/client/s3"
	"error-logging/pkg/config"
	"error-logging/services"
	serviceimpl "error-logging/services/impl"

	"go.uber.org/dig"
)

const rateLimitPerSecond = 10

func BuildWorkerContainer() *dig.Container {
	container := dig.New()

	container.Provide(config.NewDefaultConfigProvider)
	container.Provide(config.NewMysqlConfig)
	container.Provide(config.NewRedisConfig)
	container.Provide(config.NewClickhouseConfig)
	container.Provide(config.NewKafkaConfig)
	container.Provide(config.NewS3Config)
	container.Provide(config.NewWorkerConfig)

	container.Provide(mysqlclient.NewClient)
	container.Provide(provideRedis)
	container.Provide(chclient.NewClient)
	container.Provide(chclient.NewNativeClient)
	container.Provide(kafkaclient.NewClient)
	container.Provide(s3client.NewClient)

	container.Provide(repositoryimpl.NewIssueRepository)
	container.Provide(repositoryimpl.NewLogRepository)
	container.Provide(repositoryimpl.NewErrorEventRepository)

	container.Provide(func(c *s3client.Client) services.ObjectStore { return c })
	container.Provide(func(r *redisclient.Client) services.IssueCache {
		return serviceimpl.NewIssueCache(r)
	})
	container.Provide(func(r *redisclient.Client) services.RateLimiter {
		return serviceimpl.NewRateLimiter(r, rateLimitPerSecond)
	})

	container.Provide(serviceimpl.NewProcessorService)
	container.Provide(serviceimpl.NewBatchConsumer)

	return container
}
