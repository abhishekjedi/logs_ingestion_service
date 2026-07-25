package middleware

import (
	"net/http"

	"error-logging/pkg/config"
	"error-logging/pkg/session"

	"github.com/gin-gonic/gin"
)

// ContextUserIDKey is the gin-context key holding the authenticated user id.
const ContextUserIDKey = "user_id"

// AuthMiddleware authenticates dashboard requests via the session cookie.
type AuthMiddleware struct {
	session    *session.Manager
	cookieName string
}

func NewAuthMiddleware(sess *session.Manager, cfg config.AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{session: sess, cookieName: cfg.CookieName}
}

// RequireUser rejects requests without a valid session cookie, otherwise stores
// the user id in the context.
func (m *AuthMiddleware) RequireUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(m.cookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		userID, err := m.session.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			return
		}
		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}

// UserIDFromContext returns the authenticated user id set by RequireUser.
func UserIDFromContext(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}
