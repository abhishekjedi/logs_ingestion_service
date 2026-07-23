package di

import (
	controllerimpl "error-logging/controllers/impl"
	repositoryimpl "error-logging/db/repository/impl"
	"error-logging/dto"
	"error-logging/middleware"
	chclient "error-logging/pkg/client/clickhouse"
	kafkaclient "error-logging/pkg/client/kafka"
	mysqlclient "error-logging/pkg/client/mysql"
	s3client "error-logging/pkg/client/s3"
	"error-logging/pkg/config"
	"error-logging/router"
	serviceimpl "error-logging/services/impl"

	"go.uber.org/dig"
)

func BuildContainer() *dig.Container {
	container := dig.New()

	// Configs
	container.Provide(config.NewDefaultConfigProvider)
	container.Provide(config.NewAppConfig)
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

	// Repositories
	container.Provide(repositoryimpl.NewProjectRepository)
	container.Provide(repositoryimpl.NewServiceRepository)

	// Services
	container.Provide(serviceimpl.NewProjectService)
	container.Provide(serviceimpl.NewServiceService)
	container.Provide(serviceimpl.NewIngestService)

	// Middleware
	container.Provide(middleware.NewAPIKeyAuth)

	// Controllers
	container.Provide(controllerimpl.NewHealthController)
	container.Provide(controllerimpl.NewProjectController)
	container.Provide(controllerimpl.NewServiceController)
	container.Provide(controllerimpl.NewIngestController)

	// Routers
	container.Provide(router.NewHealthRouter)
	container.Provide(router.NewProjectRouter)
	container.Provide(router.NewIngestRouter)
	container.Provide(func(hr *router.HealthRouter, pr *router.ProjectRouter, ir *router.IngestRouter) *router.Router {
		return &router.Router{
			Routes: []dto.Route{hr, pr, ir},
		}
	})

	return container
}
