package router

import (
	"error-logging/controllers"
	"error-logging/middleware"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

// IssueRouter registers the issue read API (login required, access checked per resource).
type IssueRouter struct {
	IssueController controllers.IssueController
	Auth            *middleware.AuthMiddleware
}

func NewIssueRouter(ic controllers.IssueController, auth *middleware.AuthMiddleware) *IssueRouter {
	return &IssueRouter{IssueController: ic, Auth: auth}
}

func (r *IssueRouter) Register(api *gin.RouterGroup) {
	g := api.Group("")
	g.Use(r.Auth.RequireUser())

	g.GET("/services/:service_id/issues", apicontext.Wrap(r.IssueController.ListIssues))
	g.GET("/issues/:issue_id", apicontext.Wrap(r.IssueController.GetIssue))
	g.GET("/issues/:issue_id/timeseries", apicontext.Wrap(r.IssueController.GetTimeseries))
	g.GET("/issues/:issue_id/events", apicontext.Wrap(r.IssueController.GetEvents))
}
