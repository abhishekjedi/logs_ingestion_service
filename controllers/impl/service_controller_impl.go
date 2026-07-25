package impl

import (
	"net/http"

	"error-logging/controllers"
	"error-logging/dto"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

type serviceController struct {
	svc   services.ServiceService
	authz services.AuthzService
}

func NewServiceController(svc services.ServiceService, authz services.AuthzService) controllers.ServiceController {
	return &serviceController{svc: svc, authz: authz}
}

func (ctl *serviceController) CreateService(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	projectID, ok := uintParam(c, "project_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	if err := ctl.authz.RequireProjectAccess(c.Request.Context(), userID, projectID); err != nil {
		respondErr(c, err)
		return
	}

	var req dto.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := ctl.svc.CreateService(c.Request.Context(), projectID, req)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (ctl *serviceController) ListServices(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	projectID, ok := uintParam(c, "project_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}
	if err := ctl.authz.RequireProjectAccess(c.Request.Context(), userID, projectID); err != nil {
		respondErr(c, err)
		return
	}

	svcs, err := ctl.svc.ListServices(c.Request.Context(), projectID)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"services": svcs})
}
