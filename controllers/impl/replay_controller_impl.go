package impl

import (
	"net/http"
	"strings"

	"error-logging/controllers"
	"error-logging/dto"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

type replayController struct {
	service services.ReplayService
	authz   services.AuthzService
}

func NewReplayController(
	service services.ReplayService,
	authz services.AuthzService,
) controllers.ReplayController {
	return &replayController{service: service, authz: authz}
}

func (c *replayController) ListIntegrations(api *context.ApiContext) {
	userID, projectID, ok := c.projectAccess(api, false)
	if !ok {
		return
	}
	_ = userID
	integrations, err := c.service.ListIntegrations(api.Request.Context(), projectID)
	if err != nil {
		respondErr(api, err)
		return
	}
	canManage := c.authz.RequireProjectManage(api.Request.Context(), userID, projectID) == nil
	api.JSON(http.StatusOK, gin.H{"integrations": integrations, "can_manage": canManage})
}

func (c *replayController) UpsertIntegration(api *context.ApiContext) {
	_, projectID, ok := c.projectAccess(api, true)
	if !ok {
		return
	}
	var request dto.UpsertReplayIntegrationRequest
	if err := api.ShouldBindJSON(&request); err != nil {
		api.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	integration, err := c.service.UpsertIntegration(api.Request.Context(), projectID, request)
	if err != nil {
		respondErr(api, err)
		return
	}
	api.JSON(http.StatusOK, gin.H{"integration": integration})
}

func (c *replayController) DeleteIntegration(api *context.ApiContext) {
	_, projectID, ok := c.projectAccess(api, true)
	if !ok {
		return
	}
	if err := c.service.DeleteIntegration(
		api.Request.Context(),
		projectID,
		strings.TrimSpace(api.Query("project_key")),
	); err != nil {
		respondErr(api, err)
		return
	}
	api.Status(http.StatusNoContent)
}

func (c *replayController) GetSessionContext(api *context.ApiContext) {
	userID, ok := currentUser(api)
	if !ok {
		return
	}
	issueID, ok := uintParam(api, "issue_id")
	if !ok {
		api.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue id"})
		return
	}
	if err := c.authz.RequireIssueAccess(api.Request.Context(), userID, issueID); err != nil {
		respondErr(api, err)
		return
	}
	eventID := strings.TrimSpace(api.Param("event_id"))
	if eventID == "" {
		api.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
		return
	}
	response, err := c.service.GetSessionContext(api.Request.Context(), issueID, eventID)
	if err != nil {
		respondErr(api, err)
		return
	}
	api.JSON(http.StatusOK, response)
}

func (c *replayController) projectAccess(
	api *context.ApiContext,
	manage bool,
) (userID, projectID uint64, ok bool) {
	userID, ok = currentUser(api)
	if !ok {
		return 0, 0, false
	}
	projectID, ok = uintParam(api, "project_id")
	if !ok {
		api.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return 0, 0, false
	}
	var err error
	if manage {
		err = c.authz.RequireProjectManage(api.Request.Context(), userID, projectID)
	} else {
		err = c.authz.RequireProjectAccess(api.Request.Context(), userID, projectID)
	}
	if err != nil {
		respondErr(api, err)
		return 0, 0, false
	}
	return userID, projectID, true
}
