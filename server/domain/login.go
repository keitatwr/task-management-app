package domain

import (
	"context"

	"github.com/gin-gonic/gin"
)

type LoginUsecase interface {
	GetUserByEmail(context.Context, string) (*User, error)
	CreateSession(*gin.Context, User) error
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
