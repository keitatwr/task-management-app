package domain

import "context"

type TaskPermission struct {
	ID      int  `json:"id"`
	TaskID  int  `json:"taskID"`
	UserID  int  `json:"userID"`
	CanEdit bool `json:"canEdit"`
	CanRead bool `json:"canRead"`
}

type TaskPermissionRepository interface {
	GrantPermission(ctx context.Context, taskPermission *TaskPermission) error
	FetchTaskIDByUserID(ctx context.Context, id int, canEdit, canRead bool) ([]int, error)
	FetchPermissionByTaskID(ctx context.Context, taskID, userID int) (*TaskPermission, error)
	// GetPermissionByUserID(ctx context.Context, taskID, userID int) (*TaskPermission, error)
	// Update(ctx context.Context, taskPermission *TaskPermission) error
}

// type TaskPermissionUsecase interface {
// 	Create(ctx context.Context, taskPermission *TaskPermission) error
// 	GetPermissionByUserID(ctx context.Context, taskID, userID int) (*TaskPermission, error)
// 	Update(ctx context.Context, taskPermission *TaskPermission) error
// }
