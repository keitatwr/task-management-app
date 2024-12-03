package session

import (
	"encoding/json"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/domain"
)

type SessionManager interface {
	CreateSession(ctx *gin.Context, user domain.User) error
	GetSession(ctx *gin.Context) (domain.User, error)
}

type sessionManager struct{}

func NewSessionManager() SessionManager {
	return &sessionManager{}
}

func (sm *sessionManager) CreateSession(ctx *gin.Context, user domain.User) error {
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

func (sm *sessionManager) GetSession(ctx *gin.Context) (domain.User, error) {
	session := sessions.Default(ctx)
	userInfoJson := session.Get("userInfo")
	if userInfoJson == nil {
		return domain.User{}, nil
	}

	var userInfo domain.User
	if err := json.Unmarshal([]byte(userInfoJson.(string)), &userInfo); err != nil {
		return domain.User{}, err
	}
	return userInfo, nil
}
