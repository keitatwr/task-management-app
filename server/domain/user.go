package domain

import (
	"context"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthUser struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserRepository interface {
	Create(context.Context, *User) error
	GetUserByID(context.Context, int) (*User, error)
	GetUserByEmail(context.Context, string) (*User, error)
	GetAllUser(context.Context) ([]User, error)
	Delete(context.Context, int) error
}

type UserUsecases interface {
	Create(context.Context, *User) error
	GetUser(context.Context, int) (User, error)
	GetAllUser(context.Context) ([]User, error)
	Delete(context.Context, int) error
}
