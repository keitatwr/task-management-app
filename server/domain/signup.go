package domain

import "context"

type SignupUsecase interface {
	Create(ctx context.Context, name, email, password string) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}
