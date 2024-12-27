package route

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/api/controller"
	"github.com/keitatwr/task-management-app/internal/security"
	"github.com/keitatwr/task-management-app/internal/session"
	"github.com/keitatwr/task-management-app/repository"
	"github.com/keitatwr/task-management-app/usecase"
	"gorm.io/gorm"
)

func NewLoginRouter(timeout time.Duration, db *gorm.DB, r *gin.RouterGroup) {
	ur := repository.NewUserReposiotry(db)
	lc := controller.LoginController{
		LoginUsecase:      usecase.NewLoginUsecase(ur, session.NewSessionManager()),
		PasswordCompareer: &security.BcryptPasswordComparer{},
	}
	r.POST("/login", lc.Login)
}
