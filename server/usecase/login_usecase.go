package usecase

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/session"
)

type loginUsecase struct {
	userRepository domain.UserRepository
	sessionManager session.SessionManager
	contextTimeout time.Duration
}

func NewLoginUsecase(ur domain.UserRepository, sm session.SessionManager,
	timeout time.Duration) domain.LoginUsecase {
	return &loginUsecase{
		userRepository: ur,
		sessionManager: sm,
		contextTimeout: timeout,
	}
}

func (lu *loginUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := lu.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (lu *loginUsecase) CreateSession(ctx *gin.Context, user domain.User) error {
	return lu.sessionManager.CreateSession(ctx, user)
}
