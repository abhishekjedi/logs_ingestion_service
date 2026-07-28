package impl

import (
	"encoding/json"
	"io"
	"net/http"

	"error-logging/controllers"
	"error-logging/middleware"
	"error-logging/pkg/context"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

const maxIngestBytes = 10 << 20

type ingestController struct {
	svc services.IngestService
}

func NewIngestController(svc services.IngestService) controllers.IngestController {
	return &ingestController{svc: svc}
}

func (ctl *ingestController) Ingest(c *context.ApiContext) {
	service, ok := middleware.ServiceFromContext(c.Context)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not resolved"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxIngestBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
		return
	}
	if len(body) == 0 || !json.Valid(body) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body must be OTLP logs JSON"})
		return
	}

	if err := ctl.svc.Ingest(c.Request.Context(), service, body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to accept payload"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}
