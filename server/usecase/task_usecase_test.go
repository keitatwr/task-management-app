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

var AnyTime domain.DateOnly

func TestCreateTask(t *testing.T) {
	type args struct {
		title                       string
		description                 string
		userID                      int
		dueDate                     domain.DateOnly
		setupMockTaskRepo           func(*mock.MockTaskRepository)
		setupMockTaskPermissionRepo func(*mock.MockTaskPermissionRepository)
	}

	tests := []struct {
		title     string
		args      args
		wantError error
	}{
		{
			"success",
			args{
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyTime,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().Create(context.TODO(), &domain.Task{
						Title:       "test title",
						Description: "test description",
						Completed:   false,
						CreatedBy:   1,
						DueDate:     AnyTime,
					}).Return(1, nil)
				},
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().GrantPermission(context.TODO(), &domain.TaskPermission{
						TaskID:  1,
						UserID:  1,
						CanEdit: true,
						CanRead: true,
					}).Return(nil)
				},
			},
			nil,
		},
		{
			"transaction not found",
			args{
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyTime,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().Create(context.TODO(), &domain.Task{
						Title:       "test title",
						Description: "test description",
						Completed:   false,
						CreatedBy:   1,
						DueDate:     AnyTime,
					}).Return(1, myerror.ErrTransactionNotFound)
				},
				setupMockTaskPermissionRepo: nil,
			},
			myerror.ErrTransactionNotFound,
		},
		{
			"create task failed",
			args{
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyTime,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().Create(context.TODO(), &domain.Task{
						Title:       "test title",
						Description: "test description",
						Completed:   false,
						CreatedBy:   1,
						DueDate:     AnyTime,
					}).Return(0, myerror.ErrQueryFailed)
				},
				setupMockTaskPermissionRepo: nil,
			},
			myerror.ErrQueryFailed,
		},
		{
			"grant permission failed",
			args{
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyTime,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().Create(context.TODO(), &domain.Task{
						Title:       "test title",
						Description: "test description",
						Completed:   false,
						CreatedBy:   1,
						DueDate:     AnyTime,
					}).Return(1, nil)
				},
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().GrantPermission(context.TODO(), &domain.TaskPermission{
						TaskID:  1,
						UserID:  1,
						CanEdit: true,
						CanRead: true,
					}).Return(myerror.ErrPermissionDenied)
				},
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

			if tt.args.setupMockTaskRepo != nil {
				tt.args.setupMockTaskRepo(mockTaskRepo)
			}
			if tt.args.setupMockTaskPermissionRepo != nil {
				tt.args.setupMockTaskPermissionRepo(mockTaskPermissionRepo)
			}

			// run
			uc := usecase.NewTaskUsecase(mockTaskRepo, mockTaskPermissionRepo, &transaction.Noop{})
			err := uc.Create(context.TODO(), tt.args.title, tt.args.description, tt.args.userID, tt.args.dueDate)

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
		userID                      int
		setupMockTaskRepo           func(*mock.MockTaskRepository)
		setupMockTaskPermissionRepo func(*mock.MockTaskPermissionRepository)
	}

	tests := []struct {
		title     string
		args      args
		wantTasks []domain.Task
		wantError error
	}{
		{
			"success",
			args{
				userID: 1,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().FetchAllTaskByTaskID(context.TODO(), 1, 2).
						Return([]domain.Task{
							{ID: 1, Title: "Task 1"},
							{ID: 2, Title: "Task 2"},
						}, nil)
				},
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchTaskIDByUserID(context.TODO(), 1, true, true).
						Return([]int{1, 2}, nil)
				},
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
				userID:            1,
				setupMockTaskRepo: nil,
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchTaskIDByUserID(context.TODO(), 1, true, true).
						Return(nil, myerror.ErrQueryFailed)
				},
			},
			nil,
			myerror.ErrQueryFailed,
		},
		{
			"fetch tasks failed",
			args{
				userID: 1,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().FetchAllTaskByTaskID(context.TODO(), 1, 2).
						Return(nil, myerror.ErrQueryFailed)
				},
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchTaskIDByUserID(context.TODO(), 1, true, true).
						Return([]int{1, 2}, nil)
				},
			},
			nil,
			myerror.ErrQueryFailed,
		},
		{
			"fetch task permission not found",
			args{
				userID:            1,
				setupMockTaskRepo: nil,
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchTaskIDByUserID(context.TODO(), 1, true, true).
						Return(nil, myerror.ErrPermissionNotFound)
				},
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

			if tt.args.setupMockTaskRepo != nil {
				tt.args.setupMockTaskRepo(mockTaskRepo)
			}
			if tt.args.setupMockTaskPermissionRepo != nil {
				tt.args.setupMockTaskPermissionRepo(mockTaskPermissionRepo)
			}

			// run
			uc := usecase.NewTaskUsecase(mockTaskRepo, mockTaskPermissionRepo, &transaction.Noop{})
			tasks, err := uc.FetchAllTaskByUserID(context.TODO(), tt.args.userID)

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
		taskID                      int
		userID                      int
		setupMockTaskRepo           func(*mock.MockTaskRepository)
		setupMockTaskPermissionRepo func(*mock.MockTaskPermissionRepository)
	}

	tests := []struct {
		title     string
		args      args
		wantTask  *domain.Task
		wantError error
	}{
		{
			"success",
			args{
				taskID: 1,
				userID: 1,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().FetchTaskByTaskID(context.TODO(), 1).
						Return(&domain.Task{ID: 1, Title: "Task 1"}, nil)
				},
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
						Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
				},
			},
			&domain.Task{ID: 1, Title: "Task 1"},
			nil,
		},
		{
			"fetch task permission not found",
			args{
				taskID:            1,
				userID:            1,
				setupMockTaskRepo: nil,
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
						Return(nil, myerror.ErrPermissionNotFound)
				},
			},
			nil,
			myerror.ErrPermissionNotFound,
		},
		{
			"fetch task permission denied",
			args{
				taskID:            1,
				userID:            1,
				setupMockTaskRepo: nil,
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
						Return(&domain.TaskPermission{CanRead: false, CanEdit: false}, nil)
				},
			},
			nil,
			myerror.ErrPermissionDenied,
		},
		{
			"fetch task failsd",
			args{
				taskID: 1,
				userID: 1,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().FetchTaskByTaskID(context.TODO(), 1).
						Return(nil, myerror.ErrQueryFailed)
				},
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
						Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
				},
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

			if tt.args.setupMockTaskRepo != nil {
				tt.args.setupMockTaskRepo(mockTaskRepo)
			}
			if tt.args.setupMockTaskPermissionRepo != nil {
				tt.args.setupMockTaskPermissionRepo(mockTaskPermissionRepo)
			}

			// run
			uc := usecase.NewTaskUsecase(mockTaskRepo, mockTaskPermissionRepo, &transaction.Noop{})
			task, err := uc.FetchTaskByTaskID(context.TODO(), tt.args.taskID, tt.args.userID)

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
		taskID                      int
		title                       string
		description                 string
		userID                      int
		dueDate                     domain.DateOnly
		setupMockTaskRepo           func(*mock.MockTaskRepository)
		setupMockTaskPermissionRepo func(*mock.MockTaskPermissionRepository)
	}
	tests := []struct {
		title     string
		args      args
		wantError error
	}{
		{
			"success",
			args{
				taskID:      1,
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyTime,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().Update(context.TODO(), &domain.Task{
						ID:          1,
						Title:       "test title",
						Description: "test description",
						DueDate:     AnyTime,
					}).Return(nil)
				},
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
						Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
				},
			},
			nil,
		},
		{
			"update task permission not found",
			args{
				taskID:            1,
				title:             "test title",
				description:       "test description",
				userID:            1,
				dueDate:           AnyTime,
				setupMockTaskRepo: nil,
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
						Return(nil, myerror.ErrPermissionNotFound)
				},
			},
			myerror.ErrPermissionNotFound,
		},
		{
			"update task permission denied",
			args{
				taskID:            1,
				title:             "test title",
				description:       "test description",
				userID:            1,
				dueDate:           AnyTime,
				setupMockTaskRepo: nil,
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
						Return(&domain.TaskPermission{CanRead: false, CanEdit: false}, nil)
				},
			},
			myerror.ErrPermissionDenied,
		},
		{
			"update task failed",
			args{
				taskID:      1,
				title:       "test title",
				description: "test description",
				userID:      1,
				dueDate:     AnyTime,
				setupMockTaskRepo: func(mockTaskRepo *mock.MockTaskRepository) {
					mockTaskRepo.EXPECT().Update(context.TODO(), &domain.Task{
						ID:          1,
						Title:       "test title",
						Description: "test description",
						DueDate:     AnyTime,
					}).Return(myerror.ErrQueryFailed)
				},
				setupMockTaskPermissionRepo: func(mockTaskPermissionRepo *mock.MockTaskPermissionRepository) {
					mockTaskPermissionRepo.EXPECT().FetchPermissionByTaskID(context.TODO(), 1, 1).
						Return(&domain.TaskPermission{CanRead: true, CanEdit: true}, nil)
				},
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

			if tt.args.setupMockTaskRepo != nil {
				tt.args.setupMockTaskRepo(mockTaskRepo)
			}
			if tt.args.setupMockTaskPermissionRepo != nil {
				tt.args.setupMockTaskPermissionRepo(mockTaskPermissionRepo)
			}

			// run
			uc := usecase.NewTaskUsecase(mockTaskRepo, mockTaskPermissionRepo, &transaction.Noop{})
			err := uc.Update(context.TODO(), tt.args.taskID, tt.args.userID, tt.args.title, tt.args.description, tt.args.dueDate)

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
