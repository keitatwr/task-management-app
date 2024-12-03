package usecases

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

func (lu *loginUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := lu.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (lu *loginUsecase) CreateSession(ctx *gin.Context, user domain.User) error {
	session := sessions.Default(ctx)
	bUser, err := json.Marshal(user)
	if err != nil {
		return err
	}
	session.Set("userInfo", string(bUser))
	session.Options(sessions.Options{MaxAge: 3600, Path: "/", HttpOnly: false})
	if err := session.Save(); err != nil {
		return err
	}
	return nil
}
