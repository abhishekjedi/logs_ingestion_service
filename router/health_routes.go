package router

import (
	"error-logging/controllers"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

type HealthRouter struct {
	HealthController controllers.HealthController
}

func NewHealthRouter(hc controllers.HealthController) *HealthRouter {
	return &HealthRouter{HealthController: hc}
}

func (r *HealthRouter) Register(api *gin.RouterGroup) {
	healthGroup := api.Group("/health")
	healthGroup.GET("/", apicontext.Wrap(r.HealthController.GetAppHealth))
}
