package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/api/response"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
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
			err := myerror.ErrNoLogin.WithDescription("user not logged in")
			response.Error(c, http.StatusUnauthorized, "unauthorized", err)
			c.Abort()
			return
		}
		var authUser domain.User
		if err := json.Unmarshal([]byte(authUserJson.(string)), &authUser); err != nil {
			err := myerror.ErrUnExpected.WrapWithDescription(err, "occurrred unexpected error")
			response.Error(c, http.StatusInternalServerError, "unexpected error", err)
			c.Abort()
			return
		}
		SetUserContext(c, authUser)
		c.Next()
	}
}
