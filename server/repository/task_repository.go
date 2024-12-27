package repository

import (
	"context"
	"errors"

	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"gorm.io/gorm"
)

var GetTxFunc = GetTx

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) domain.TaskRepository {
	return &taskRepository{
		db: db,
	}
}

func (r *taskRepository) Create(ctx context.Context, task *domain.Task) (int, error) {
	tx, ok := GetTxFunc(ctx)
	if !ok {
		return -1, myerror.ErrTransactionNotFound
	}
	if err := tx.Create(task).Error; err != nil {
		return -1, myerror.ErrQueryFailed.Wrap(err)
	}
	return task.ID, nil
}

func (r *taskRepository) FetchAllTaskByTaskID(ctx context.Context, taskIDs ...int) ([]domain.Task, error) {
	var tasks []domain.Task
	if err := r.db.WithContext(ctx).Where("id IN ?", taskIDs).Find(&tasks).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerror.ErrTaskNotFound.Wrap(err)
		}
		return nil, myerror.ErrQueryFailed.Wrap(err)
	}
	return tasks, nil
}

func (r *taskRepository) FetchTaskByTaskID(ctx context.Context, taskID int) (*domain.Task, error) {
	var task domain.Task
	if err := r.db.WithContext(ctx).Where("id = ?", taskID).Take(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerror.ErrTaskNotFound.Wrap(err)
		}
		return nil, myerror.ErrQueryFailed.Wrap(err)
	}
	return &task, nil
}

func (r *taskRepository) Update(ctx context.Context, taskID int, updateFields map[string]any) error {
	var task domain.Task
	if err := r.db.WithContext(ctx).Model(&task).Where("id = ?", taskID).Select("title", "description", "due_date").Updates(updateFields).Error; err != nil {
		return myerror.ErrQueryFailed.Wrap(err)
	}
	return nil
}

func (r *taskRepository) Delete(ctx context.Context, taskID int) error {
	if err := r.db.WithContext(ctx).Where("id = ?", taskID).Delete(&domain.Task{}).Error; err != nil {
		return myerror.ErrQueryFailed.Wrap(err)
	}
	return nil
}
