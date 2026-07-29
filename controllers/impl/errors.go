package impl

import (
	"errors"
	"net/http"

	"error-logging/middleware"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

func respondErr(c *context.ApiContext, err error) {
	switch {
	case errors.Is(err, services.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, services.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, services.ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrUpstream):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func currentUser(c *context.ApiContext) (uint64, bool) {
	id, ok := middleware.UserIDFromContext(c.Context)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return 0, false
	}
	return id, true
}
