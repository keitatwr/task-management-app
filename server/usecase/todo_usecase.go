package usecase

import (
	"context"
	"time"

	"github.com/keitatwr/todo-app/domain"
)

type todoUsecase struct {
	todoRepository domain.TodoRepository
	timeout        time.Duration
}

func NewTodoUsecase(tr domain.TodoRepository, timeout time.Duration) domain.TodoUsecase {
	return &todoUsecase{
		todoRepository: tr,
		timeout:        timeout,
	}
}

func (tu *todoUsecase) Create(ctx context.Context, title, description string, userID int) error {
	todo := &domain.Todo{
		Title:       title,
		Description: description,
		Completed:   false,
		UserID:      userID,
	}
	err := tu.todoRepository.Create(ctx, todo)
	if err != nil {
		return err
	}
	return nil
}

func (tu *todoUsecase) GetTodoByID(ctx context.Context, id int) (*domain.Todo, error) {
	return tu.todoRepository.GetTodoByID(ctx, id)
}

func (tu *todoUsecase) GetAllTodoByUserID(ctx context.Context, userID int) ([]domain.Todo, error) {
	return tu.todoRepository.GetAllTodoByUserID(ctx, userID)
}

func (tu *todoUsecase) Update(ctx context.Context, id int) error {
	return tu.todoRepository.Update(ctx, id)
}

func (tu *todoUsecase) Delete(ctx context.Context, id int) error {
	return tu.todoRepository.Delete(ctx, id)
}
