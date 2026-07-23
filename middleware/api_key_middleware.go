// Package middleware holds cross-cutting gin handlers.
package middleware

import (
	"net/http"

	dbdto "error-logging/db/dto"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

// ContextServiceKey is the gin-context key under which the authenticated service
// is stored for downstream handlers.
const ContextServiceKey = "service"

// APIKeyAuth authenticates ingest requests via the X-API-Key header.
type APIKeyAuth struct {
	svc services.ServiceService
}

func NewAPIKeyAuth(svc services.ServiceService) *APIKeyAuth {
	return &APIKeyAuth{svc: svc}
}

// Handle resolves the API key to a service, verifies it matches the service in the
// URL path, and stores the service in the context. Aborts 401/403 otherwise.
func (a *APIKeyAuth) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing api key"})
			return
		}

		service, err := a.svc.AuthenticateKey(c.Request.Context(), key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		// The key must belong to the service addressed in the URL.
		if pid := c.Param("service_public_id"); pid != "" && service.PublicID != pid {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "api key does not match service"})
			return
		}

		c.Set(ContextServiceKey, service)
		c.Next()
	}
}

// ServiceFromContext returns the authenticated service placed by Handle.
func ServiceFromContext(c *gin.Context) (*dbdto.Service, bool) {
	v, ok := c.Get(ContextServiceKey)
	if !ok {
		return nil, false
	}
	service, ok := v.(*dbdto.Service)
	return service, ok
}
