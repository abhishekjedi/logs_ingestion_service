package router

import (
	"error-logging/controllers"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

// IssueRouter registers the issue read API.
type IssueRouter struct {
	IssueController controllers.IssueController
}

func NewIssueRouter(ic controllers.IssueController) *IssueRouter {
	return &IssueRouter{IssueController: ic}
}

func (r *IssueRouter) Register(api *gin.RouterGroup) {
	api.GET("/services/:service_id/issues", apicontext.Wrap(r.IssueController.ListIssues))
	api.GET("/issues/:issue_id", apicontext.Wrap(r.IssueController.GetIssue))
	api.GET("/issues/:issue_id/timeseries", apicontext.Wrap(r.IssueController.GetTimeseries))
	api.GET("/issues/:issue_id/events", apicontext.Wrap(r.IssueController.GetEvents))
}
