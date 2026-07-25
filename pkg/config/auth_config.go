package config

import (
	"time"

	"github.com/knadh/koanf/v2"
)

// AuthConfig configures dashboard (user) auth: the JWT session cookie and
// Google OAuth.
type AuthConfig struct {
	JWTSecret    string
	TokenTTL     time.Duration
	CookieName   string
	CookieSecure bool // set true behind HTTPS

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

// GoogleEnabled reports whether Google OAuth is fully configured.
func (a AuthConfig) GoogleEnabled() bool {
	return a.GoogleClientID != "" && a.GoogleClientSecret != "" && a.GoogleRedirectURL != ""
}

func NewAuthConfig(k *koanf.Koanf) AuthConfig {
	secret := k.String("auth.jwt_secret")
	if secret == "" {
		secret = "dev-insecure-secret-change-me"
	}
	cookie := k.String("auth.cookie_name")
	if cookie == "" {
		cookie = "errlog_session"
	}
	return AuthConfig{
		JWTSecret:    secret,
		TokenTTL:     kDuration(k, "auth.token_ttl", 24*time.Hour),
		CookieName:   cookie,
		CookieSecure: k.Bool("auth.cookie_secure"),

		GoogleClientID:     k.String("auth.google_client_id"),
		GoogleClientSecret: k.String("auth.google_client_secret"),
		GoogleRedirectURL:  k.String("auth.google_redirect_url"),
	}
}
