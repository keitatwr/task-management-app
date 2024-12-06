package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/keitatwr/todo-app/domain"
	"github.com/keitatwr/todo-app/tests/mocks"
	"github.com/keitatwr/todo-app/usecase"
	"github.com/stretchr/testify/assert"
)

func getMockRepository(t *testing.T) (*mocks.MockTodoRepository, func()) {
	mockCtrl := gomock.NewController(t)
	tearDown := func() {
		mockCtrl.Finish()
	}
	return mocks.NewMockTodoRepository(mockCtrl), tearDown
}

func TestCreateTodo(t *testing.T) {
	type args struct {
		title       string
		description string
		userID      int
	}
	tests := []struct {
		title         string
		args          args
		expectedError bool
	}{
		{
			"create todo successfully",
			args{
				title:       "test",
				description: "test",
				userID:      1,
			},
			false,
		},
		{
			"create todo with error",
			args{
				title:       "test",
				description: "test",
				userID:      1,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockRepo, tearDown := getMockRepository(t)
			defer tearDown()
			if tt.expectedError {
				mockRepo.EXPECT().Create(context.TODO(), &domain.Todo{
					Title:       tt.args.title,
					Description: tt.args.description,
					UserID:      tt.args.userID,
				}).Return(fmt.Errorf("error"))
			} else {
				mockRepo.EXPECT().Create(context.TODO(), &domain.Todo{
					Title:       tt.args.title,
					Description: tt.args.description,
					UserID:      tt.args.userID,
				}).Return(nil)
			}

			// run
			uc := usecase.NewTodoUsecase(mockRepo, 0)
			err := uc.Create(context.TODO(), tt.args.title, tt.args.description,
				tt.args.userID)

			// assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTodoByID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		title         string
		args          int
		expected      *domain.Todo
		expectedError bool
	}{
		{
			"create todo successfully",
			1,
			&domain.Todo{
				ID:          1,
				Title:       "test",
				Description: "test",
				Completed:   false,
				UserID:      1,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			false,
		},
		{
			"create todo with error",
			1,
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockRepo, tearDown := getMockRepository(t)
			defer tearDown()
			if tt.expectedError {
				mockRepo.EXPECT().GetTodoByID(context.TODO(), 1).Return(nil, fmt.Errorf("error"))
			} else {
				mockRepo.EXPECT().GetTodoByID(context.TODO(), 1).Return(&domain.Todo{
					ID:          1,
					Title:       "test",
					Description: "test",
					Completed:   false,
					UserID:      1,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
			}

			// run
			uc := usecase.NewTodoUsecase(mockRepo, 0)
			todo, err := uc.GetTodoByID(context.TODO(), tt.args)

			// assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, todo)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, todo)
			}
		})
	}
}

func TestGetAllTodoByUserID(t *testing.T) {
	now := time.Now()
	tests := []struct {
		title         string
		args          int
		expected      []domain.Todo
		expectedError bool
	}{
		{
			"create todo successfully",
			1,
			[]domain.Todo{
				{
					ID:          1,
					Title:       "test1",
					Description: "test1",
					Completed:   false,
					UserID:      1,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				{
					ID:          2,
					Title:       "test2",
					Description: "test2",
					Completed:   false,
					UserID:      1,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
			false,
		},
		{
			"create todo with error",
			1,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockRepo, tearDown := getMockRepository(t)
			defer tearDown()
			if tt.expectedError {
				mockRepo.EXPECT().GetAllTodoByUserID(context.TODO(), 1).Return(nil, fmt.Errorf("error"))
			} else {
				mockRepo.EXPECT().GetAllTodoByUserID(context.TODO(), 1).
					Return([]domain.Todo{
						{
							ID:          1,
							Title:       "test1",
							Description: "test1",
							Completed:   false,
							UserID:      1,
							CreatedAt:   now,
							UpdatedAt:   now,
						},
						{
							ID:          2,
							Title:       "test2",
							Description: "test2",
							Completed:   false,
							UserID:      1,
							CreatedAt:   now,
							UpdatedAt:   now,
						},
					}, nil)
			}

			// run
			uc := usecase.NewTodoUsecase(mockRepo, 0)
			todos, err := uc.GetAllTodoByUserID(context.TODO(), tt.args)

			// assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expected, todos)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, todos)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		title         string
		args          int
		expectedError bool
	}{
		{
			"update todo successfully",
			1,
			false,
		},
		{
			"update todo with error",
			1,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockRepo, tearDown := getMockRepository(t)
			defer tearDown()
			if tt.expectedError {
				mockRepo.EXPECT().Update(context.TODO(), 1).Return(fmt.Errorf("error"))
			} else {
				mockRepo.EXPECT().Update(context.TODO(), 1).Return(nil)
			}

			// run
			uc := usecase.NewTodoUsecase(mockRepo, 0)
			err := uc.Update(context.TODO(), tt.args)

			// assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		title         string
		args          int
		expectedError bool
	}{
		{
			"delete todo successfully",
			1,
			false,
		},
		{
			"delete todo with error",
			1,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockRepo, tearDown := getMockRepository(t)
			defer tearDown()
			if tt.expectedError {
				mockRepo.EXPECT().Delete(context.TODO(), 1).Return(fmt.Errorf("error"))
			} else {
				mockRepo.EXPECT().Delete(context.TODO(), 1).Return(nil)
			}

			// run
			uc := usecase.NewTodoUsecase(mockRepo, 0)
			err := uc.Delete(context.TODO(), tt.args)

			// assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
