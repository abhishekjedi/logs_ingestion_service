package impl

import (
	"net/http"

	"error-logging/controllers"
	"error-logging/db/repository"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

type issueController struct {
	svc services.IssueService
}

func NewIssueController(svc services.IssueService) controllers.IssueController {
	return &issueController{svc: svc}
}

func (ctl *issueController) ListIssues(c *context.ApiContext) {
	serviceID, ok := uintParam(c, "service_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}

	limit := queryInt(c, "limit", 50)
	offset := queryInt(c, "offset", 0)
	filter := repository.IssueListFilter{
		ServiceID: serviceID,
		Status:    c.Query("status"),
		Sort:      c.Query("sort"),
		Order:     c.Query("order"),
		Limit:     limit,
		Offset:    offset,
	}

	res, err := ctl.svc.ListIssues(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	id, ok := uintParam(c, "issue_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue id"})
		return
	}
	issue, err := ctl.svc.GetIssue(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "issue not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"issue": issue})
}

func (ctl *issueController) GetTimeseries(c *context.ApiContext) {
	id, ok := uintParam(c, "issue_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue id"})
		return
	}
	from, to := parseTimeRange(c)
	points, err := ctl.svc.GetTimeseries(c.Request.Context(), id, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func (ctl *issueController) GetEvents(c *context.ApiContext) {
	id, ok := uintParam(c, "issue_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid issue id"})
		return
	}
	events, err := ctl.svc.GetEvents(c.Request.Context(), id, queryInt(c, "limit", 50))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}
