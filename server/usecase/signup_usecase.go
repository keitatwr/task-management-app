package usecase

import (
	"context"

	"github.com/keitatwr/task-management-app/domain"
)

type signupUsecase struct {
	userRepository domain.UserRepository
}

func NewSignupUsecase(ur domain.UserRepository) domain.SignupUsecase {
	return &signupUsecase{
		userRepository: ur,
	}
}

func (su signupUsecase) Create(ctx context.Context, name, email, password string) error {
	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: password,
	}
	err := su.userRepository.Create(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (su signupUsecase) FetchUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := su.userRepository.FetchUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}
