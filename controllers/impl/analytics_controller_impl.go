package impl

import (
	"net/http"
	"time"

	"error-logging/controllers"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

type analyticsController struct {
	svc   services.AnalyticsService
	authz services.AuthzService
}

func NewAnalyticsController(svc services.AnalyticsService, authz services.AuthzService) controllers.AnalyticsController {
	return &analyticsController{svc: svc, authz: authz}
}

func (ctl *analyticsController) GetServiceOverview(c *context.ApiContext) {
	serviceID, ok := ctl.serviceAccess(c)
	if !ok {
		return
	}
	from, to := parseTimeRange(c)
	points, err := ctl.svc.GetServiceOverview(c.Request.Context(), serviceID, from, to)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"points": points})
}

func (ctl *analyticsController) GetReleaseHealth(c *context.ApiContext) {
	serviceID, ok := ctl.serviceAccess(c)
	if !ok {
		return
	}
	from, to := parseTimeRange(c)
	releases, err := ctl.svc.GetReleaseHealth(c.Request.Context(), serviceID, from, to)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"releases": releases})
}

func (ctl *analyticsController) GetBreadcrumbs(c *context.ApiContext) {
	serviceID, ok := ctl.serviceAccess(c)
	if !ok {
		return
	}
	sessionID := c.Query("session_id")
	before := time.Now().UTC()
	if v := c.Query("before"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			before = t
		}
	}
	limit := queryInt(c, "limit", 50)
	crumbs, err := ctl.svc.GetBreadcrumbs(c.Request.Context(), serviceID, sessionID, before, limit)
	if err != nil {
		respondErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"breadcrumbs": crumbs})
}

func (ctl *analyticsController) serviceAccess(c *context.ApiContext) (serviceID uint64, ok bool) {
	userID, ok := currentUser(c)
	if !ok {
		return 0, false
	}
	serviceID, ok = uintParam(c, "service_id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service id"})
		return 0, false
	}
	if err := ctl.authz.RequireServiceAccess(c.Request.Context(), userID, serviceID); err != nil {
		respondErr(c, err)
		return 0, false
	}
	return serviceID, true
}
