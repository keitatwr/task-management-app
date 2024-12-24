package route

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/api/controller"
	"github.com/keitatwr/task-management-app/repository"
	"github.com/keitatwr/task-management-app/usecase"
	"gorm.io/gorm"
)

func NewTaskRouter(timeout time.Duration, db *gorm.DB, r *gin.RouterGroup) {
	tRepo := repository.NewTaskRepository(db)
	tpRepo := repository.NewTaskPermissionRepository(db)
	transaction := repository.NewTransaction(db)
	tc := controller.TaskController{
		TaskUsecase: usecase.NewTaskUsecase(tRepo, tpRepo, transaction),
	}
	r.POST("/tasks", tc.Create)
	r.GET("/tasks", tc.FetchAllTaskByUserID)
	r.GET("/tasks/:taskID", tc.FetchTaskByTaskID)
	// r.PUT("/todo", tc.Update)
	// r.DELETE("/todo", tc.Delete)
}
