package usecase_test

import (
	"context"
	"testing"

	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/tests/mock"
	"github.com/keitatwr/task-management-app/transaction"
	"github.com/keitatwr/task-management-app/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func getMockTaskRepository(mockCtrl *gomock.Controller) *mock.MockTaskRepository {

	return mock.NewMockTaskRepository(mockCtrl)
}

func getMockTaskPermissionRepository(mockCtrl *gomock.Controller) *mock.MockTaskPermissionRepository {

	return mock.NewMockTaskPermissionRepository(mockCtrl)
}

var AnyDate domain.DateOnly

func TestCreateTask(t *testing.T) {
	type args struct {
		ctx         context.Context
		title       string
		description string
		userID      int
		dueDate     domain.DateOnly
	}

	tests := []struct {
		title                       string
		args                        args
		setupMockTaskRepo           func(*mock.MockTaskRepository)
		setupMockTaskPermissionRepo func(*mock.MockTaskPermissionRepository)
		wantError                   error
	}{
		{
			"success",
			args{
				ctx:         context.TODO(),
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyDate,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().Create(context.TODO(), &domain.Task{
					Title:       "test title",
					Description: "test description",
					Completed:   false,
					CreatedBy:   1,
					DueDate:     AnyDate,
				}).Return(1, nil)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().GrantPermission(context.TODO(), &domain.TaskPermission{
					TaskID:  1,
					UserID:  1,
					CanEdit: true,
					CanRead: true,
				}).Return(nil)
			},
			nil,
		},
		{
			"transaction not found",
			args{
				ctx:         context.TODO(),
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyDate,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().Create(context.TODO(), &domain.Task{
					Title:       "test title",
					Description: "test description",
					Completed:   false,
					CreatedBy:   1,
					DueDate:     AnyDate,
				}).Return(1, myerror.ErrTransactionNotFound)
			},
			nil,
			myerror.ErrTransactionNotFound,
		},
		{
			"create task failed",
			args{
				ctx:         context.TODO(),
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyDate,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().Create(context.TODO(), &domain.Task{
					Title:       "test title",
					Description: "test description",
					Completed:   false,
					CreatedBy:   1,
					DueDate:     AnyDate,
				}).Return(0, myerror.ErrQueryFailed)
			},
			nil,
			myerror.ErrQueryFailed,
		},
		{
			"grant permission failed",
			args{
				ctx:         context.TODO(),
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyDate,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().Create(context.TODO(), &domain.Task{
					Title:       "test title",
					Description: "test description",
					Completed:   false,
					CreatedBy:   1,
					DueDate:     AnyDate,
				}).Return(1, nil)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().GrantPermission(context.TODO(), &domain.TaskPermission{
					TaskID:  1,
					UserID:  1,
					CanEdit: true,
					CanRead: true,
				}).Return(myerror.ErrPermissionDenied)
			},
			myerror.ErrPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockTaskRepo := getMockTaskRepository(ctrl)
			mockTaskPermissionRepo := getMockTaskPermissionRepository(ctrl)

			if tt.setupMockTaskRepo != nil {
				tt.setupMockTaskRepo(mockTaskRepo)
			}
			if tt.setupMockTaskPermissionRepo != nil {
				tt.setupMockTaskPermissionRepo(mockTaskPermissionRepo)
			}

			// run
			uc := usecase.NewTaskUsecase(mockTaskRepo, mockTaskPermissionRepo, &transaction.Noop{})
			err := uc.Create(tt.args.ctx, tt.args.title, tt.args.description, tt.args.userID, tt.args.dueDate)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFetchAllTaskByUserID(t *testing.T) {
	type args struct {
		ctx    context.Context
		userID int
	}

	tests := []struct {
		title                       string
		args                        args
		setupMockTaskRepo           func(*mock.MockTaskRepository)
		setupMockTaskPermissionRepo func(*mock.MockTaskPermissionRepository)
		wantTasks                   []domain.Task
		wantError                   error
	}{
		{
			"success",
			args{
				ctx:    context.TODO(),
				userID: 1,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().FetchAllTaskByTaskID(context.TODO(), 1, 2).
					Return([]domain.Task{
						{ID: 1, Title: "Task 1"},
						{ID: 2, Title: "Task 2"},
					}, nil)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchTaskIDByUserID(context.TODO(), 1, true, true).
					Return([]int{1, 2}, nil)
			},
			[]domain.Task{
				{ID: 1, Title: "Task 1"},
				{ID: 2, Title: "Task 2"},
			},
			nil,
		},
		{
			"fetch task IDs failed",
			args{
				ctx:    context.TODO(),
				userID: 1,
			},
			nil,
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchTaskIDByUserID(context.TODO(), 1, true, true).
					Return(nil, myerror.ErrQueryFailed)
			},
			nil,
			myerror.ErrQueryFailed,
		},
		{
			"fetch tasks failed",
			args{
				ctx:    context.TODO(),
				userID: 1,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().FetchAllTaskByTaskID(context.TODO(), 1, 2).
					Return(nil, myerror.ErrQueryFailed)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchTaskIDByUserID(context.TODO(), 1, true, true).
					Return([]int{1, 2}, nil)
			},
			nil,
			myerror.ErrQueryFailed,
		},
		{
			"fetch task permission not found",
			args{
				ctx:    context.TODO(),
				userID: 1,
			},
			nil,
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchTaskIDByUserID(context.TODO(), 1, true, true).
					Return(nil, myerror.ErrPermissionNotFound)
			},
			nil,
			myerror.ErrPermissionNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockTaskRepo := getMockTaskRepository(ctrl)
			mockTaskPermissionRepo := getMockTaskPermissionRepository(ctrl)

			if tt.setupMockTaskRepo != nil {
				tt.setupMockTaskRepo(mockTaskRepo)
			}
			if tt.setupMockTaskPermissionRepo != nil {
				tt.setupMockTaskPermissionRepo(mockTaskPermissionRepo)
			}

			// run
			uc := usecase.NewTaskUsecase(mockTaskRepo, mockTaskPermissionRepo, &transaction.Noop{})
			tasks, err := uc.FetchAllTaskByUserID(tt.args.ctx, tt.args.userID)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTasks, tasks)
			}
		})
	}
}

func TestFetchTaskByTaskID(t *testing.T) {
	type args struct {
		ctx    context.Context
		taskID int
		userID int
	}

	tests := []struct {
		title                       string
		args                        args
		setupMockTaskRepo           func(*mock.MockTaskRepository)
		setupMockTaskPermissionRepo func(*mock.MockTaskPermissionRepository)
		wantTask                    *domain.Task
		wantError                   error
	}{
		{
			"success",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().FetchTaskByTaskID(context.TODO(), 1).
					Return(&domain.Task{ID: 1, Title: "Task 1"}, nil)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
			},
			&domain.Task{ID: 1, Title: "Task 1"},
			nil,
		},
		{
			"fetch task permission not found",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			nil,
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(nil, myerror.ErrPermissionNotFound)
			},
			nil,
			myerror.ErrPermissionNotFound,
		},
		{
			"fetch task permission denied",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			nil,
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(&domain.TaskPermission{CanRead: false, CanEdit: false}, nil)
			},
			nil,
			myerror.ErrPermissionDenied,
		},
		{
			"fetch task failsd",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().FetchTaskByTaskID(context.TODO(), 1).
					Return(nil, myerror.ErrQueryFailed)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
			},
			nil,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockTaskRepo := getMockTaskRepository(ctrl)
			mockTaskPermissionRepo := getMockTaskPermissionRepository(ctrl)

			if tt.setupMockTaskRepo != nil {
				tt.setupMockTaskRepo(mockTaskRepo)
			}
			if tt.setupMockTaskPermissionRepo != nil {
				tt.setupMockTaskPermissionRepo(mockTaskPermissionRepo)
			}

			// run
			uc := usecase.NewTaskUsecase(mockTaskRepo, mockTaskPermissionRepo, &transaction.Noop{})
			task, err := uc.FetchTaskByTaskID(tt.args.ctx, tt.args.taskID, tt.args.userID)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTask, task)
			}
		})
	}
}

func TestUpdateTask(t *testing.T) {
	type args struct {
		ctx         context.Context
		taskID      int
		title       string
		description string
		userID      int
		dueDate     domain.DateOnly
	}

	tests := []struct {
		title                       string
		args                        args
		setupMockTaskRepo           func(*mock.MockTaskRepository)
		setupMockTaskPermissionRepo func(*mock.MockTaskPermissionRepository)
		wantError                   error
	}{
		{
			"success",
			args{
				ctx:         context.TODO(),
				taskID:      1,
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyDate,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().Update(context.TODO(), 1, map[string]any{
					"title":       "test title",
					"description": "test description",
					"due_date":    AnyDate,
				}).Return(nil)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
			},
			nil,
		},
		{
			"update task permission not found",
			args{
				ctx:         context.TODO(),
				taskID:      1,
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyDate,
			},
			nil,
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(nil, myerror.ErrPermissionNotFound)
			},
			myerror.ErrPermissionNotFound,
		},
		{
			"update task permission denied",
			args{
				ctx:         context.TODO(),
				taskID:      1,
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyDate,
			},
			nil,
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(&domain.TaskPermission{CanRead: false, CanEdit: false}, nil)
			},
			myerror.ErrPermissionDenied,
		},
		{
			"update task failed",
			args{
				ctx:         context.TODO(),
				taskID:      1,
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyDate,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().Update(context.TODO(), 1, map[string]any{
					"title":       "test title",
					"description": "test description",
					"due_date":    AnyDate,
				}).Return(myerror.ErrQueryFailed)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
			},
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockTaskRepo := getMockTaskRepository(ctrl)
			mockTaskPermissionRepo := getMockTaskPermissionRepository(ctrl)

			if tt.setupMockTaskRepo != nil {
				tt.setupMockTaskRepo(mockTaskRepo)
			}
			if tt.setupMockTaskPermissionRepo != nil {
				tt.setupMockTaskPermissionRepo(mockTaskPermissionRepo)
			}

			// run
			uc := usecase.NewTaskUsecase(mockTaskRepo, mockTaskPermissionRepo, &transaction.Noop{})
			err := uc.Update(tt.args.ctx, tt.args.taskID, tt.args.userID, tt.args.title, tt.args.description, tt.args.dueDate)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteTask(t *testing.T) {
	type args struct {
		ctx    context.Context
		taskID int
		userID int
	}

	tests := []struct {
		title                       string
		args                        args
		setupMockTaskRepo           func(*mock.MockTaskRepository)
		setupMockTaskPermissionRepo func(*mock.MockTaskPermissionRepository)
		wantError                   error
	}{
		{
			"success",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().Delete(context.TODO(), 1).Return(nil)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
			},
			nil,
		},
		{
			"delete task permission not found",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			nil,
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(nil, myerror.ErrPermissionNotFound)
			},
			myerror.ErrPermissionNotFound,
		},
		{
			"delete task permission denied",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			nil,
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(&domain.TaskPermission{CanRead: false, CanEdit: false}, nil)
			},
			myerror.ErrPermissionDenied,
		},
		{
			"delete task failed",
			args{
				ctx:    context.TODO(),
				taskID: 1,
				userID: 1,
			},
			func(mockTaskRepo *mock.MockTaskRepository) {
				mockTaskRepo.EXPECT().Delete(context.TODO(), 1).Return(myerror.ErrQueryFailed)
			},
			func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
				mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
					Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
			},
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockTaskRepo := getMockTaskRepository(ctrl)
			mockTaskPermissionRepo := getMockTaskPermissionRepository(ctrl)

			if tt.setupMockTaskRepo != nil {
				tt.setupMockTaskRepo(mockTaskRepo)
			}
			if tt.setupMockTaskPermissionRepo != nil {
				tt.setupMockTaskPermissionRepo(mockTaskPermissionRepo)
			}

			// run
			uc := usecase.NewTaskUsecase(mockTaskRepo, mockTaskPermissionRepo, &transaction.Noop{})
			err := uc.Delete(tt.args.ctx, tt.args.taskID, tt.args.userID)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
