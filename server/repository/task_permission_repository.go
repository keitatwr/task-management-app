package repository

import (
	"context"
	"errors"

	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"gorm.io/gorm"
)

type taskPermissionRepository struct {
	db *gorm.DB
}

func NewTaskPermissionRepository(db *gorm.DB) domain.TaskPermissionRepository {
	return &taskPermissionRepository{
		db: db,
	}
}

func (r *taskPermissionRepository) GrantPermission(ctx context.Context, taskPermission *domain.TaskPermission) error {
	tx, ok := GetTxFunc(ctx)
	if !ok {
		return myerror.ErrTransactionNotFound
	}
	if err := tx.Create(taskPermission).Error; err != nil {
		return myerror.ErrGrantPermission.Wrap(err)
	}
	return nil
}

func (r *taskPermissionRepository) FetchTaskIDByUserID(ctx context.Context, userID int, canEdit, canRead bool) ([]int, error) {
	var taskIDs []int
	var taskPermission domain.TaskPermission
	query := r.db.WithContext(ctx).Model(&taskPermission).Select("task_id").Where("user_id = ?", userID)

	query = query.Where("can_edit = ? AND can_read = ?", canEdit, canRead)

	if err := query.Find(&taskIDs).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerror.ErrPermissionNotFound.Wrap(err)
		}
		return nil, myerror.ErrQueryFailed.Wrap(err)
	}
	return taskIDs, nil
}

func (r *taskPermissionRepository) FetchPermissionByTaskID(ctx context.Context, taskID int, userID int) (*domain.TaskPermission, error) {
	var taskPermission domain.TaskPermission
	if err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Where("user_id = ?", userID).Take(&taskPermission).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, myerror.ErrPermissionNotFound.Wrap(err)
		}
		return nil, myerror.ErrQueryFailed.Wrap(err)
	}
	return &taskPermission, nil
}
