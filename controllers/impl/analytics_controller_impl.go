package impl

import (
	"net/http"

	"error-logging/controllers"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

type analyticsController struct {
	svc services.AnalyticsService
}

func NewAnalyticsController(svc services.AnalyticsService) controllers.AnalyticsController {
	return &analyticsController{svc: svc}
}

func (ctl *analyticsController) GetServiceOverview(c *context.ApiContext) {
	serviceID, ok := uintParam(c, "service_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}
	from, to := parseTimeRange(c)
	points, err := ctl.svc.GetServiceOverview(c.Request.Context(), serviceID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func (ctl *analyticsController) GetReleaseHealth(c *context.ApiContext) {
	serviceID, ok := uintParam(c, "service_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return
	}
	from, to := parseTimeRange(c)
	releases, err := ctl.svc.GetReleaseHealth(c.Request.Context(), serviceID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"releases": releases})
}
