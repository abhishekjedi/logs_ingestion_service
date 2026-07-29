package router

import (
	"error-logging/controllers"
	"error-logging/middleware"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

type ReplayRouter struct {
	Controller controllers.ReplayController
	Auth       *middleware.AuthMiddleware
}

func NewReplayRouter(controller controllers.ReplayController, auth *middleware.AuthMiddleware) *ReplayRouter {
	return &ReplayRouter{Controller: controller, Auth: auth}
}

func (r *ReplayRouter) Register(api *gin.RouterGroup) {
	group := api.Group("")
	group.Use(r.Auth.RequireUser())

	group.GET("/projects/:project_id/integrations/openreplay", apicontext.Wrap(r.Controller.ListIntegrations))
	group.PUT("/projects/:project_id/integrations/openreplay", apicontext.Wrap(r.Controller.UpsertIntegration))
	group.DELETE("/projects/:project_id/integrations/openreplay", apicontext.Wrap(r.Controller.DeleteIntegration))
	group.GET("/issues/:issue_id/events/:event_id/session-context", apicontext.Wrap(r.Controller.GetSessionContext))
}
