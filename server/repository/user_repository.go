package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserReposiotry(db *gorm.DB) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (ur *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := ur.db.WithContext(ctx).Create(user).Error; err != nil {
		return myerror.ErrQueryFailed.Wrap(err)
	}
	return nil
}

func (ur *userRepository) FetchUserByID(ctx context.Context, id int) (*domain.User, error) {
	// var user domain.User
	// if err := ur.db.WithContext(ctx).Where("id = ?", id).Take(&user).Error; err != nil {
	// 	return nil, err
	// }
	// return &user, nil
	return nil, fmt.Errorf("not implemented yet")
}

func (ur *userRepository) FetchUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := ur.db.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerror.ErrUserNotFound.Wrap(err)
		}
		return nil, myerror.ErrQueryFailed.Wrap(err)
	}
	return &user, nil
}

func (ur *userRepository) Delete(ctx context.Context, id int) error {
	return ur.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&domain.User{}).Error; err != nil {
			return myerror.ErrQueryFailed.Wrap(err)
		}
		return nil
	})
}
