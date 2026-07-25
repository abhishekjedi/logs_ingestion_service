package impl

import (
	"context"

	dbdto "error-logging/db/dto"
	"error-logging/db/repository"
	mysqlclient "error-logging/pkg/client/mysql"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(c *mysqlclient.Client) repository.UserRepository {
	return &userRepository{db: c.DB}
}

func (r *userRepository) FindOrCreateByEmail(ctx context.Context, email, name string) (*dbdto.User, error) {
	var user dbdto.User
	err := r.db.WithContext(ctx).
		Where(dbdto.User{Email: email}).
		Attrs(dbdto.User{Name: name}).
		FirstOrCreate(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id uint64) (*dbdto.User, error) {
	var user dbdto.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*dbdto.User, error) {
	var user dbdto.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
