package repository

import (
	"context"

	"github.com/keitatwr/todo-app/domain"
	"gorm.io/gorm"
)

type todoRepository struct {
	db *gorm.DB
}

func NewTodoRepository(db *gorm.DB) domain.TodoRepository {
	return &todoRepository{
		db: db,
	}
}

func (tr *todoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	return tr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(todo).Error; err != nil {
			return err
		}
		return nil
	})
}

func (tr *todoRepository) GetTodoByID(ctx context.Context, id int) (*domain.Todo, error) {
	var todo domain.Todo
	if err := tr.db.WithContext(ctx).Where("id = ?", id).Take(&todo).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}

func (tr *todoRepository) GetAllTodoByUserID(ctx context.Context, userID int) ([]domain.Todo, error) {
	var todos []domain.Todo
	if err := tr.db.WithContext(ctx).Where("user_id = ?", userID).Find(&todos).Error; err != nil {
		return nil, err
	}
	return todos, nil
}

func (tr *todoRepository) Update(ctx context.Context, id int) error {
	var todo *domain.Todo
	return tr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(todo).Where("id = ?", id).Update("completed", true).Error; err != nil {
			return err
		}
		return nil
	})
}

func (tr *todoRepository) Delete(ctx context.Context, id int) error {
	return tr.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).Delete(&domain.Todo{}).Error; err != nil {
			return err
		}
		return nil
	})
}
