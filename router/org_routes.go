package router

import (
	"error-logging/controllers"
	"error-logging/middleware"
	apicontext "error-logging/pkg/context"

	"github.com/gin-gonic/gin"
)

type OrgRouter struct {
	OrgController controllers.OrgController
	Auth          *middleware.AuthMiddleware
}

func NewOrgRouter(oc controllers.OrgController, auth *middleware.AuthMiddleware) *OrgRouter {
	return &OrgRouter{OrgController: oc, Auth: auth}
}

func (r *OrgRouter) Register(api *gin.RouterGroup) {
	orgs := api.Group("/orgs")
	orgs.Use(r.Auth.RequireUser())

	orgs.POST("", apicontext.Wrap(r.OrgController.CreateOrg))
	orgs.GET("", apicontext.Wrap(r.OrgController.ListMyOrgs))
	orgs.POST("/:org_id/members", apicontext.Wrap(r.OrgController.InviteMember))
	orgs.GET("/:org_id/members", apicontext.Wrap(r.OrgController.ListMembers))
	orgs.DELETE("/:org_id/members/:member_id", apicontext.Wrap(r.OrgController.RemoveMember))
}
