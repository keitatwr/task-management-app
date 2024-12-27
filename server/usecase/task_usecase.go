package usecase

import (
	"context"

	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/transaction"
)

type taskUsecase struct {
	taskRepository           domain.TaskRepository
	taskPermissionRepository domain.TaskPermissionRepository
	transaction              transaction.Transaction
}

func NewTaskUsecase(taskRepo domain.TaskRepository,
	taskPermissionRepo domain.TaskPermissionRepository,
	transaction transaction.Transaction) domain.TaskUsecase {
	return &taskUsecase{
		taskRepository:           taskRepo,
		taskPermissionRepository: taskPermissionRepo,
		transaction:              transaction,
	}
}

func (u *taskUsecase) Create(ctx context.Context,
	title, description string, userID int, dueDate domain.DateOnly) error {
	_, err := u.transaction.DoInTx(ctx, func(ctx context.Context) (interface{}, error) {
		todo := &domain.Task{
			Title:       title,
			Description: description,
			Completed:   false,
			CreatedBy:   userID,
			DueDate:     dueDate,
		}
		todoID, err := u.taskRepository.Create(ctx, todo)
		if err != nil {
			return nil, err
		}

		taskPermission := &domain.TaskPermission{
			TaskID:  todoID,
			UserID:  userID,
			CanEdit: true,
			CanRead: true,
		}
		err = u.taskPermissionRepository.GrantPermission(ctx, taskPermission)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (u *taskUsecase) FetchAllTaskByUserID(ctx context.Context, userID int) ([]domain.Task, error) {
	taskIDs, err := u.taskPermissionRepository.FetchTaskIDByUserID(ctx, userID, true, true)
	if err != nil {
		return nil, err
	}
	tasks, err := u.taskRepository.FetchAllTaskByTaskID(ctx, taskIDs...)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (u *taskUsecase) FetchTaskByTaskID(ctx context.Context, taskID, userID int) (*domain.Task, error) {
	permisison, err := u.taskPermissionRepository.FetchPermissionByTaskID(ctx, taskID, userID)
	if err != nil {
		return nil, err
	}
	if !permisison.CanRead && !permisison.CanEdit {
		return nil, myerror.ErrPermissionDenied
	}

	task, err := u.taskRepository.FetchTaskByTaskID(ctx, taskID)
	if err != nil {
		// logger.E(ctx, "failed to fetch task by taskID: %v", err)
		return nil, err
	}
	return task, nil
}

func (u *taskUsecase) Update(ctx context.Context, taskID, userID int, title, description string, dueDate domain.DateOnly) error {
	permisison, err := u.taskPermissionRepository.FetchPermissionByTaskID(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if !permisison.CanRead && !permisison.CanEdit {
		return myerror.ErrPermissionDenied
	}

	update_fileds := map[string]any{
		"title":       title,
		"description": description,
		"due_date":    dueDate,
	}

	return u.taskRepository.Update(ctx,
		taskID, update_fileds)
}

func (u *taskUsecase) Delete(ctx context.Context, taskID, userID int) error {
	permisison, err := u.taskPermissionRepository.FetchPermissionByTaskID(ctx, taskID, userID)
	if err != nil {
		return err
	}
	if !permisison.CanEdit {
		return myerror.ErrPermissionDenied
	}

	return u.taskRepository.Delete(ctx, taskID)
}
