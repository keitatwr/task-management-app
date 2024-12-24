package route

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/api/middleware"
	"gorm.io/gorm"
)

func Setup(timeout time.Duration, db *gorm.DB, r *gin.Engine) {
	r.Use(gin.Recovery())
	r.Use(middleware.LoggingMiddleware(
		middleware.NewLoggerConfig(
			middleware.WithBaseLogLevel(slog.LevelInfo),
			middleware.WithClientErrorLogLevel(slog.LevelWarn),
			middleware.WithServerErrorLogLevel(slog.LevelError),
		)))
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("sessionid", store))
	publicRouter := r.Group("")
	NewSignupRouter(timeout, db, publicRouter)
	NewLoginRouter(timeout, db, publicRouter)
	privateRouter := r.Group("")
	privateRouter.Use(middleware.AuthMiddleware())
	NewTaskRouter(timeout, db, privateRouter)
}
