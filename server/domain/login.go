package domain

import (
	"context"

	"github.com/gin-gonic/gin"
)

type LoginUsecase interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateSession(ctx *gin.Context, user User) error
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
