package repository

import (
	"context"

	"github.com/keitatwr/todo-app/domain"
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
	return ur.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return nil
	})
}

func (ur *userRepository) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	var user domain.User
	if err := ur.db.WithContext(ctx).Where("id = ?", id).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := ur.db.WithContext(ctx).Where("email = ?", email).Take(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *userRepository) GetAllUser(ctx context.Context) ([]domain.User, error) {
	var users []domain.User
	if err := ur.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (ur *userRepository) Delete(ctx context.Context, id int) error {
	return ur.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&domain.User{}).Error; err != nil {
			return err
		}
		return nil
	})
}
