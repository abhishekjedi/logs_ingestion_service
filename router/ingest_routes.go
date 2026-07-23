package router

import (
	"error-logging/controllers"
	"error-logging/middleware"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

// IngestRouter registers the OTLP ingest endpoint, guarded by API-key auth.
type IngestRouter struct {
	IngestController controllers.IngestController
	Auth             *middleware.APIKeyAuth
}

func NewIngestRouter(ic controllers.IngestController, auth *middleware.APIKeyAuth) *IngestRouter {
	return &IngestRouter{IngestController: ic, Auth: auth}
}

func (r *IngestRouter) Register(api *gin.RouterGroup) {
	logs := api.Group("/v1/logs")
	logs.Use(r.Auth.Handle())
	logs.POST("/:service_public_id", apicontext.Wrap(r.IngestController.Ingest))
}
