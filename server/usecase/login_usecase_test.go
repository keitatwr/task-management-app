package usecase_test

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/keitatwr/todo-app/domain"
	"github.com/keitatwr/todo-app/tests/mocks"
	"github.com/keitatwr/todo-app/usecase"
	"github.com/stretchr/testify/assert"
)

func getMockSessionManager(t *testing.T) (*mocks.MockSessionManager, func()) {
	mockCtrl := gomock.NewController(t)
	tearDown := func() {
		defer mockCtrl.Finish()
	}
	return mocks.NewMockSessionManager(mockCtrl), tearDown
}

func TestCreateSession(t *testing.T) {
	type args struct {
		userID int
	}
	tests := []struct {
		title         string
		args          domain.User
		expectedError bool
	}{
		{
			"create session successfully",
			domain.User{
				ID:       1,
				Name:     "test name",
				Email:    "test email",
				Password: "test password",
			},
			false,
		},
		{
			"fail to create session",
			domain.User{
				ID:       1,
				Name:     "test name",
				Email:    "test email",
				Password: "test password",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mock, tearDown := getMockSessionManager(t)
			defer tearDown()

			// Create a new gin context
			ginCtx, _ := gin.CreateTestContext(nil)

			if tt.expectedError {
				mock.EXPECT().CreateSession(ginCtx, tt.args).Return(fmt.Errorf("failed to create session"))
			} else {
				mock.EXPECT().CreateSession(ginCtx, tt.args).Return(nil)
			}

			lu := usecase.NewLoginUsecase(nil, mock, 0)
			err := lu.CreateSession(ginCtx, tt.args)

			// assert
			if tt.expectedError {
				assert.Equal(t, "failed to create session", err.Error())
			} else {
				assert.Equal(t, nil, err)

			}

		})
	}
}
