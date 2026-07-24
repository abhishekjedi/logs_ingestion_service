package router

import (
	"error-logging/controllers"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

// AnalyticsRouter registers the service-level analytics read API.
type AnalyticsRouter struct {
	AnalyticsController controllers.AnalyticsController
}

func NewAnalyticsRouter(ac controllers.AnalyticsController) *AnalyticsRouter {
	return &AnalyticsRouter{AnalyticsController: ac}
}

func (r *AnalyticsRouter) Register(api *gin.RouterGroup) {
	api.GET("/services/:service_id/overview", apicontext.Wrap(r.AnalyticsController.GetServiceOverview))
	api.GET("/services/:service_id/release-health", apicontext.Wrap(r.AnalyticsController.GetReleaseHealth))
}
