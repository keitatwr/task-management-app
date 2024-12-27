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

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FetchUserByID(ctx context.Context, id int) (*User, error)
	FetchUserByEmail(ctx context.Context, email string) (*User, error)
	Delete(ctx context.Context, id int) error
}
