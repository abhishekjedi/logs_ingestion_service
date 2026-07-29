package middleware

import (
	"net/http"

	"error-logging/pkg/config"

	"github.com/gin-gonic/gin"
)

type CORS struct {
	allowedOrigin string
}

func NewCORS(cfg config.AppConfig) *CORS {
	return &CORS{allowedOrigin: cfg.FrontendURL}
}

func (m *CORS) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && (origin == m.allowedOrigin || m.allowedOrigin == "*") {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header(
				"Access-Control-Allow-Headers",
				"Content-Type, X-API-Key, X-OpenReplay-Session-ID, X-OpenReplay-Project-Key, X-OpenReplay-Session-URL",
			)
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
