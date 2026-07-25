package impl

import (
	"net/http"

	"error-logging/constants"
	"error-logging/controllers"
	dbdto "error-logging/db/dto"
	"error-logging/dto"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

type orgController struct {
	svc services.OrgService
}

func NewOrgController(svc services.OrgService) controllers.OrgController {
	return &orgController{svc: svc}
}

func (ctl *orgController) CreateOrg(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.CreateOrgRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	org, err := ctl.svc.CreateOrg(c.Request.Context(), userID, req.Name)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"organization": org})
}

func (ctl *orgController) ListMyOrgs(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	orgs, err := ctl.svc.ListMyOrgs(c.Request.Context(), userID)
	if err != nil {
		respondErr(c, err)
		return
	}
	if orgs == nil {
		orgs = []dbdto.Organization{}
	}
	c.JSON(http.StatusOK, gin.H{"organizations": orgs})
}

func (ctl *orgController) InviteMember(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	orgID, ok := uintParam(c, "org_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	var req dto.InviteMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	member, err := ctl.svc.InviteMember(c.Request.Context(), userID, orgID, req.Email, constants.OrgRole(req.Role))
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"member": member})
}

func (ctl *orgController) ListMembers(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	orgID, ok := uintParam(c, "org_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	members, err := ctl.svc.ListMembers(c.Request.Context(), userID, orgID)
	if err != nil {
		respondErr(c, err)
		return
	}
	if members == nil {
		members = []dbdto.OrganizationMember{}
	}
	c.JSON(http.StatusOK, gin.H{"members": members})
}

func (ctl *orgController) RemoveMember(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	orgID, ok := uintParam(c, "org_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	memberID, ok := uintParam(c, "member_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid member id"})
		return
	}
	if err := ctl.svc.RemoveMember(c.Request.Context(), userID, orgID, memberID); err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}
