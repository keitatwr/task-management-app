package route

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/api/controller"
	"github.com/keitatwr/todo-app/internal/security"
	"github.com/keitatwr/todo-app/internal/session"
	"github.com/keitatwr/todo-app/repository"
	"github.com/keitatwr/todo-app/usecase"
	"gorm.io/gorm"
)

func NewLoginRouter(timeout time.Duration, db *gorm.DB, r *gin.RouterGroup) {
	ur := repository.NewUserReposiotry(db)
	lc := controller.LoginController{
		LoginUsecase:      usecase.NewLoginUsecase(ur, session.NewSessionManager(), timeout),
		PasswordCompareer: &security.BcryptPasswordComparer{},
	}
	r.POST("/login", lc.Login)
}
