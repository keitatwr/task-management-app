package usecases

import (
	"context"
	"time"

	"github.com/keitatwr/todo-app/domain"
)

type loginUsecase struct {
	userRepository domain.UserRepository
	contextTimeout time.Duration
}

func NewLoginUsecase(ur domain.UserRepository, timeout time.Duration) domain.LoginUsecase {
	return &loginUsecase{
		userRepository: ur,
		contextTimeout: timeout,
	}
}

func (lu loginUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := lu.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}
