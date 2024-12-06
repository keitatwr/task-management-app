package route

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/api/controller"
	"github.com/keitatwr/todo-app/repository"
	usecases "github.com/keitatwr/todo-app/usecase"
	"gorm.io/gorm"
)

func NewTodoRouter(timeout time.Duration, db *gorm.DB, r *gin.RouterGroup) {
	ur := repository.NewTodoRepository(db)
	tc := controller.TodoController{
		TodoUsecase: usecases.NewTodoUsecase(ur, timeout),
	}
	r.POST("/todo", tc.Create)
	r.GET("/todo", tc.GetAllTodoByUserID)
	r.PUT("/todo", tc.Update)
	r.DELETE("/todo", tc.Delete)
}
