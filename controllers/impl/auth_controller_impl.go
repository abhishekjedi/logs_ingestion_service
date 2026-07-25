package impl

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"

	"error-logging/controllers"
	"error-logging/middleware"
	"error-logging/pkg/config"
	"error-logging/pkg/context"
	"error-logging/pkg/session"
	"error-logging/services"

	"github.com/gin-gonic/gin"
)

const oauthStateCookie = "errlog_oauth_state"

type authController struct {
	svc    services.AuthService
	cfg    config.AuthConfig
	appCfg config.AppConfig
	sess   *session.Manager
}

func NewAuthController(svc services.AuthService, cfg config.AuthConfig, appCfg config.AppConfig, sess *session.Manager) controllers.AuthController {
	return &authController{svc: svc, cfg: cfg, appCfg: appCfg, sess: sess}
}

// GoogleLogin starts the OAuth handshake: sets a CSRF state cookie and redirects
// the browser to Google's consent screen.
func (ctl *authController) GoogleLogin(c *context.ApiContext) {
	if !ctl.svc.GoogleEnabled() {
		c.JSON(http.StatusNotFound, gin.H{"error": "google login not configured"})
		return
	}
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start login"})
		return
	}
	// Short-lived, httpOnly state cookie for CSRF protection on the callback.
	c.SetCookie(oauthStateCookie, state, 600, "/", "", ctl.cfg.CookieSecure, true)
	c.Redirect(http.StatusFound, ctl.svc.GoogleAuthURL(state))
}

// GoogleCallback completes the handshake: verifies state, exchanges the code for a
// verified user, sets the session cookie, and redirects back to the dashboard.
func (ctl *authController) GoogleCallback(c *context.ApiContext) {
	if !ctl.svc.GoogleEnabled() {
		c.JSON(http.StatusNotFound, gin.H{"error": "google login not configured"})
		return
	}

	// CSRF: the state in the query must match the one we stored in the cookie.
	stateCookie, _ := c.Cookie(oauthStateCookie)
	c.SetCookie(oauthStateCookie, "", -1, "/", "", ctl.cfg.CookieSecure, true) // consume it
	if stateCookie == "" || c.Query("state") != stateCookie {
		ctl.redirectLoginError(c, "invalid_state")
		return
	}
	if errParam := c.Query("error"); errParam != "" {
		ctl.redirectLoginError(c, errParam)
		return
	}
	code := c.Query("code")
	if code == "" {
		ctl.redirectLoginError(c, "missing_code")
		return
	}

	user, token, err := ctl.svc.GoogleLogin(c.Request.Context(), code)
	if err != nil {
		ctl.redirectLoginError(c, "auth_failed")
		return
	}
	_ = user

	ctl.setSessionCookie(c, token, ctl.sess.TTLSeconds())
	c.Redirect(http.StatusFound, ctl.appCfg.FrontendURL)
}

func (ctl *authController) redirectLoginError(c *context.ApiContext, reason string) {
	c.Redirect(http.StatusFound, ctl.appCfg.FrontendURL+"/login?error="+url.QueryEscape(reason))
}

func randomState() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (ctl *authController) Me(c *context.ApiContext) {
	userID, ok := middleware.UserIDFromContext(c.Context)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	user, err := ctl.svc.CurrentUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (ctl *authController) Logout(c *context.ApiContext) {
	ctl.setSessionCookie(c, "", -1) // expire the cookie
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

func (ctl *authController) setSessionCookie(c *context.ApiContext, token string, maxAge int) {
	c.SetCookie(ctl.cfg.CookieName, token, maxAge, "/", "", ctl.cfg.CookieSecure, true)
}
