package middleware

import (
	"context"
	"encoding/json"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/domain"
)

var userContextKey = struct{}{}

func SetUserContext(c *gin.Context, user domain.User) {
	ctx := context.WithValue(c.Request.Context(), userContextKey, user)
	c.Request = c.Request.WithContext(ctx)
}

func GetUserContext(c *gin.Context) *domain.User {
	user, ok := c.Request.Context().Value(userContextKey).(domain.User)
	if !ok {
		return nil
	}
	return &user
}

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
		SetUserContext(c, authUser)
		c.Next()
	}
}
