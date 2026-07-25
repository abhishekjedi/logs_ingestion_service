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
	"error-logging/pkg/session"
	"error-logging/router"
	serviceimpl "error-logging/services/impl"

	"go.uber.org/dig"
)

func BuildContainer() *dig.Container {
	container := dig.New()

	// Configs
	container.Provide(config.NewDefaultConfigProvider)
	container.Provide(config.NewAppConfig)
	container.Provide(config.NewAuthConfig)
	container.Provide(config.NewMysqlConfig)
	container.Provide(config.NewRedisConfig)
	container.Provide(config.NewClickhouseConfig)
	container.Provide(config.NewKafkaConfig)
	container.Provide(config.NewS3Config)

	// Clients (mysql/clickhouse/kafka required — errors fail startup; redis degradable)
	container.Provide(mysqlclient.NewClient)
	container.Provide(provideRedis)
	container.Provide(chclient.NewClient)
	container.Provide(chclient.NewNativeClient)
	container.Provide(kafkaclient.NewClient)
	container.Provide(s3client.NewClient)

	// Session
	container.Provide(session.NewManager)

	// Repositories
	container.Provide(repositoryimpl.NewUserRepository)
	container.Provide(repositoryimpl.NewOrgRepository)
	container.Provide(repositoryimpl.NewOrgMemberRepository)
	container.Provide(repositoryimpl.NewProjectRepository)
	container.Provide(repositoryimpl.NewServiceRepository)
	container.Provide(repositoryimpl.NewIssueRepository)
	container.Provide(repositoryimpl.NewAnalyticsRepository)

	// Services
	container.Provide(serviceimpl.NewAuthService)
	container.Provide(serviceimpl.NewAuthzService)
	container.Provide(serviceimpl.NewOrgService)
	container.Provide(serviceimpl.NewProjectService)
	container.Provide(serviceimpl.NewServiceService)
	container.Provide(serviceimpl.NewIngestService)
	container.Provide(serviceimpl.NewIssueService)
	container.Provide(serviceimpl.NewAnalyticsService)

	// Middleware
	container.Provide(middleware.NewAPIKeyAuth)
	container.Provide(middleware.NewAuthMiddleware)
	container.Provide(middleware.NewCORS)

	// Controllers
	container.Provide(controllerimpl.NewHealthController)
	container.Provide(controllerimpl.NewAuthController)
	container.Provide(controllerimpl.NewOrgController)
	container.Provide(controllerimpl.NewProjectController)
	container.Provide(controllerimpl.NewServiceController)
	container.Provide(controllerimpl.NewIngestController)
	container.Provide(controllerimpl.NewIssueController)
	container.Provide(controllerimpl.NewAnalyticsController)

	// Routers
	container.Provide(router.NewHealthRouter)
	container.Provide(router.NewAuthRouter)
	container.Provide(router.NewOrgRouter)
	container.Provide(router.NewProjectRouter)
	container.Provide(router.NewIngestRouter)
	container.Provide(router.NewIssueRouter)
	container.Provide(router.NewAnalyticsRouter)
	container.Provide(func(
		health *router.HealthRouter,
		auth *router.AuthRouter,
		org *router.OrgRouter,
		project *router.ProjectRouter,
		ingest *router.IngestRouter,
		issue *router.IssueRouter,
		analytics *router.AnalyticsRouter,
		cors *middleware.CORS,
	) *router.Router {
		return &router.Router{
			Routes: []dto.Route{health, auth, org, project, ingest, issue, analytics},
			CORS:   cors.Handler(),
		}
	})

	return container
}
