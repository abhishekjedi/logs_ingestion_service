package router

import (
	"error-logging/controllers"
	"error-logging/middleware"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

// AnalyticsRouter registers service-level analytics (login required, access checked).
type AnalyticsRouter struct {
	AnalyticsController controllers.AnalyticsController
	Auth                *middleware.AuthMiddleware
}

func NewAnalyticsRouter(ac controllers.AnalyticsController, auth *middleware.AuthMiddleware) *AnalyticsRouter {
	return &AnalyticsRouter{AnalyticsController: ac, Auth: auth}
}

func (r *AnalyticsRouter) Register(api *gin.RouterGroup) {
	g := api.Group("")
	g.Use(r.Auth.RequireUser())

	g.GET("/services/:service_id/overview", apicontext.Wrap(r.AnalyticsController.GetServiceOverview))
	g.GET("/services/:service_id/release-health", apicontext.Wrap(r.AnalyticsController.GetReleaseHealth))
	g.GET("/services/:service_id/breadcrumbs", apicontext.Wrap(r.AnalyticsController.GetBreadcrumbs))
}
