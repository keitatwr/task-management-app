package route

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/api/controller"
	"github.com/keitatwr/task-management-app/internal/security"
	"github.com/keitatwr/task-management-app/repository"
	usecases "github.com/keitatwr/task-management-app/usecase"
	"gorm.io/gorm"
)

func NewSignupRouter(timeout time.Duration, db *gorm.DB, r *gin.RouterGroup) {
	ur := repository.NewUserReposiotry(db)
	sc := controller.SignupController{
		SignupUsecase:  usecases.NewSignupUsecase(ur, timeout),
		PasswordHasher: &security.BcryptPasswordHasher{},
	}
	r.POST("/signup", sc.Signup)
}
