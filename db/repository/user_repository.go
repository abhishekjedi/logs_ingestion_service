package repository

import (
	"context"

	dbdto "error-logging/db/dto"
)

type UserRepository interface {
	FindOrCreateByEmail(ctx context.Context, email, name string) (*dbdto.User, error)
	GetByID(ctx context.Context, id uint64) (*dbdto.User, error)
	GetByEmail(ctx context.Context, email string) (*dbdto.User, error)
}
