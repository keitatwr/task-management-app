package middleware

import (
	"encoding/json"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/domain"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		authUserJson := session.Get("userInfo")
		if authUserJson == nil {
			c.JSON(401, domain.ErrorResponse{Message: "unauthorized"})
			c.Abort()
			return
		}
		var authUser domain.User
		if err := json.Unmarshal([]byte(authUserJson.(string)), &authUser); err != nil {
			c.JSON(401, domain.ErrorResponse{Message: "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
