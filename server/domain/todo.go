package domain

import (
	"context"
	"time"
)

type Todo struct {
	ID          int       `json:"id"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	UserID      int       `json:"userId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type TodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	GetTodoByID(ctx context.Context, id int) (*Todo, error)
	GetAllTodoByUserID(ctx context.Context, id int) ([]Todo, error)
	Update(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
}

type TodoUsecase interface {
	Create(ctx context.Context, title string, description string, userID int) error
	GetTodoByID(ctx context.Context, id int) (*Todo, error)
	GetAllTodoByUserID(ctx context.Context, id int) ([]Todo, error)
	Update(ctx context.Context, id int) error
	Delete(ctx context.Context, id int) error
}
