package domain

import "context"

type LoginUsecase interface {
	GetUserByEmail(context.Context, string) (*User, error)
}
