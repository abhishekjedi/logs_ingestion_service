package services

import (
	"context"

	dbdto "error-logging/db/dto"
)

type AuthService interface {
	CurrentUser(ctx context.Context, userID uint64) (*dbdto.User, error)

	GoogleEnabled() bool

	GoogleAuthURL(state string) string

	GoogleLogin(ctx context.Context, code string) (user *dbdto.User, token string, err error)
}
