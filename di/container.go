package di

import (
	"error-logging/controllers/impl"
	"error-logging/dto"
	chclient "error-logging/pkg/client/clickhouse"
	mysqlclient "error-logging/pkg/client/mysql"
	redisclient "error-logging/pkg/client/redis"
	"error-logging/pkg/config"
	"error-logging/router"

	"go.uber.org/dig"
)

func BuildContainer() *dig.Container {
	container := dig.New()

	// Configs
	container.Provide(config.NewDefaultConfigProvider)
	container.Provide(config.NewMysqlConfig)
	container.Provide(config.NewRedisConfig)
	container.Provide(config.NewClickhouseConfig)

	// Clients
	container.Provide(mysqlclient.NewClient)
	container.Provide(redisclient.NewClient)
	container.Provide(chclient.NewClient)

	// Repos

	// Services

	// Controllers
	container.Provide(impl.NewHealthController)

	// Routers
	container.Provide(router.NewHealthRouter)
	container.Provide(func(hr *router.HealthRouter) *router.Router {
		return &router.Router{
			Routes: []dto.Route{hr},
		}
	})

	return container
}
