package router

import (
	"error-logging/controllers"
	"error-logging/middleware"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

// ProjectRouter registers project routes (org-nested create/list) and the service
// routes nested under a project. All require login; access is checked per resource.
type ProjectRouter struct {
	ProjectController controllers.ProjectController
	ServiceController controllers.ServiceController
	Auth              *middleware.AuthMiddleware
}

func NewProjectRouter(pc controllers.ProjectController, sc controllers.ServiceController, auth *middleware.AuthMiddleware) *ProjectRouter {
	return &ProjectRouter{ProjectController: pc, ServiceController: sc, Auth: auth}
}

func (r *ProjectRouter) Register(api *gin.RouterGroup) {
	g := api.Group("")
	g.Use(r.Auth.RequireUser())

	g.POST("/orgs/:org_id/projects", apicontext.Wrap(r.ProjectController.CreateProject))
	g.GET("/orgs/:org_id/projects", apicontext.Wrap(r.ProjectController.ListProjects))
	g.GET("/projects/:project_id", apicontext.Wrap(r.ProjectController.GetProject))
	g.POST("/projects/:project_id/services", apicontext.Wrap(r.ServiceController.CreateService))
	g.GET("/projects/:project_id/services", apicontext.Wrap(r.ServiceController.ListServices))
}
