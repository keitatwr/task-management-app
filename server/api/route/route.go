package route

import (
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/api/middleware"
	"gorm.io/gorm"
)

func Setup(timeout time.Duration, db *gorm.DB, r *gin.Engine) {
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("sessionid", store))
	publicRouter := r.Group("")
	NewSignupRouter(timeout, db, publicRouter)
	NewLoginRouter(timeout, db, publicRouter)
	privateRouter := r.Group("")
	privateRouter.Use(middleware.AuthMiddleware())
	NewTodoRouter(timeout, db, privateRouter)
}
