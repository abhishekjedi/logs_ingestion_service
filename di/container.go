package di

import (
	controllerimpl "error-logging/controllers/impl"
	repositoryimpl "error-logging/db/repository/impl"
	"error-logging/dto"
	chclient "error-logging/pkg/client/clickhouse"
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
	container.Provide(config.NewS3Config)

	// Clients (mysql/clickhouse required — errors fail startup; redis degradable)
	container.Provide(mysqlclient.NewClient)
	container.Provide(provideRedis)
	container.Provide(chclient.NewClient)
	container.Provide(s3client.NewClient)

	// Repositories
	container.Provide(repositoryimpl.NewProjectRepository)
	container.Provide(repositoryimpl.NewServiceRepository)

	// Services
	container.Provide(serviceimpl.NewProjectService)
	container.Provide(serviceimpl.NewServiceService)

	// Controllers
	container.Provide(controllerimpl.NewHealthController)
	container.Provide(controllerimpl.NewProjectController)
	container.Provide(controllerimpl.NewServiceController)

	// Routers
	container.Provide(router.NewHealthRouter)
	container.Provide(router.NewProjectRouter)
	container.Provide(func(hr *router.HealthRouter, pr *router.ProjectRouter) *router.Router {
		return &router.Router{
			Routes: []dto.Route{hr, pr},
		}
	})

	return container
}
