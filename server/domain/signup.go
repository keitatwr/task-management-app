package domain

import "context"

type SignupRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
type SignupUsecase interface {
	Create(ctx context.Context, name, email, password string) error
	FetchUserByEmail(ctx context.Context, email string) (*User, error)
}
