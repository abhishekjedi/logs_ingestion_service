package impl

import (
	"error-logging/controllers"
	"error-logging/pkg/context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthControllerImpl struct{}

func NewHealthController() controllers.HealthController {
	return &HealthControllerImpl{}
}

func (h *HealthControllerImpl) GetAppHealth(c *context.ApiContext) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}
