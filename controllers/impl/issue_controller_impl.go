package impl

import (
	"net/http"

	"error-logging/controllers"
	"error-logging/dto"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

type issueController struct {
	svc   services.IssueService
	authz services.AuthzService
}

func NewIssueController(svc services.IssueService, authz services.AuthzService) controllers.IssueController {
	return &issueController{svc: svc, authz: authz}
}

func (ctl *issueController) ListIssues(c *context.ApiContext) {
	userID, ok := currentUser(c)
	if !ok {
		return
	}
	serviceID, ok := uintParam(c, "service_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}
	if err := ctl.authz.RequireServiceAccess(c.Request.Context(), userID, serviceID); err != nil {
		respondErr(c, err)
		return
	}

	limit := queryInt(c, "limit", 50)
	offset := queryInt(c, "offset", 0)
	filter := dto.IssueListFilter{
		ServiceID: serviceID,
		Status:    c.Query("status"),
		Sort:      c.Query("sort"),
		Order:     c.Query("order"),
		Limit:     limit,
		Offset:    offset,
	}

	res, err := ctl.svc.ListIssues(c.Request.Context(), filter)
	if err != nil {
		respondErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"issues": res.Issues,
		"total":  res.Total,
		"limit":  limit,
		"offset": offset,
	})
}

func (ctl *issueController) GetIssue(c *context.ApiContext) {
	userID, id, ok := ctl.issueAccess(c)
	if !ok {
		return
	}
	issue, err := ctl.svc.GetIssue(c.Request.Context(), id)
	if err != nil {
		respondErr(c, err)
		return
	}
	_ = userID
	c.JSON(http.StatusOK, gin.H{"issue": issue})
}

func (ctl *issueController) GetTimeseries(c *context.ApiContext) {
	_, id, ok := ctl.issueAccess(c)
	if !ok {
		return
	}
	from, to := parseTimeRange(c)
	points, err := ctl.svc.GetTimeseries(c.Request.Context(), id, from, to)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func (ctl *issueController) GetEvents(c *context.ApiContext) {
	_, id, ok := ctl.issueAccess(c)
	if !ok {
		return
	}
	events, err := ctl.svc.GetEvents(c.Request.Context(), id, queryInt(c, "limit", 50))
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (ctl *issueController) issueAccess(c *context.ApiContext) (userID, issueID uint64, ok bool) {
	userID, ok = currentUser(c)
	if !ok {
		return 0, 0, false
	}
	issueID, ok = uintParam(c, "issue_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue id"})
		return 0, 0, false
	}
	if err := ctl.authz.RequireIssueAccess(c.Request.Context(), userID, issueID); err != nil {
		respondErr(c, err)
		return 0, 0, false
	}
	return userID, issueID, true
}
