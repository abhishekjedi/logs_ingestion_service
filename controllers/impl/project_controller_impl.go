package impl

import (
	"net/http"

	"error-logging/controllers"
	"error-logging/dto"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

type projectController struct {
	svc   services.ProjectService
	authz services.AuthzService
}

func NewProjectController(svc services.ProjectService, authz services.AuthzService) controllers.ProjectController {
	return &projectController{svc: svc, authz: authz}
}

func (ctl *projectController) CreateProject(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	orgID, ok := uintParam(c, "org_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	if err := ctl.authz.RequireMember(c.Request.Context(), userID, orgID); err != nil {
		respondErr(c, err)
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := ctl.svc.CreateProject(c.Request.Context(), orgID, userID, req)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"project": project})
}

func (ctl *projectController) ListProjects(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	orgID, ok := uintParam(c, "org_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid org id"})
		return
	}
	if err := ctl.authz.RequireMember(c.Request.Context(), userID, orgID); err != nil {
		respondErr(c, err)
		return
	}

	projects, err := ctl.svc.ListProjects(c.Request.Context(), orgID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"projects": projects})
}

func (ctl *projectController) GetProject(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := uintParam(c, "project_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	if err := ctl.authz.RequireProjectAccess(c.Request.Context(), userID, id); err != nil {
		respondErr(c, err)
		return
	}

	project, err := ctl.svc.GetProject(c.Request.Context(), id)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"project": project})
}
