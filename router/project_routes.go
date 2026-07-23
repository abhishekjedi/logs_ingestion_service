package router

import (
	"error-logging/controllers"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

// ProjectRouter registers project routes and the service routes nested under a
// project (services are always addressed relative to their parent project).
type ProjectRouter struct {
	ProjectController controllers.ProjectController
	ServiceController controllers.ServiceController
}

func NewProjectRouter(pc controllers.ProjectController, sc controllers.ServiceController) *ProjectRouter {
	return &ProjectRouter{ProjectController: pc, ServiceController: sc}
}

func (r *ProjectRouter) Register(api *gin.RouterGroup) {
	projects := api.Group("/projects")

	projects.POST("", apicontext.Wrap(r.ProjectController.CreateProject))
	projects.GET("", apicontext.Wrap(r.ProjectController.ListProjects))
	projects.GET("/:project_id", apicontext.Wrap(r.ProjectController.GetProject))

	projects.POST("/:project_id/services", apicontext.Wrap(r.ServiceController.CreateService))
	projects.GET("/:project_id/services", apicontext.Wrap(r.ServiceController.ListServices))
}
