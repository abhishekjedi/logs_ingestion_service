package services

import (
	"context"

	dbdto "error-logging/db/dto"
)

// AuthService establishes a logged-in user (via Google OAuth) and mints the
// session token, ending in the "resolved user → issue token" step.
type AuthService interface {
	CurrentUser(ctx context.Context, userID uint64) (*dbdto.User, error)

	// GoogleEnabled reports whether Google OAuth is configured.
	GoogleEnabled() bool
	// GoogleAuthURL builds the consent-screen URL for the given CSRF state.
	GoogleAuthURL(state string) string
	// GoogleLogin exchanges an authorization code for a verified user + session token.
	GoogleLogin(ctx context.Context, code string) (user *dbdto.User, token string, err error)
}
