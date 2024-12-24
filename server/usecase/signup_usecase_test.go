package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/tests/mocks"
	"github.com/keitatwr/task-management-app/usecase"
	"github.com/stretchr/testify/require"
)

func getMockUserRepository(t *testing.T) (*mocks.MockUserRepository, func()) {
	mockCtrl := gomock.NewController(t)
	tearDown := func() {
		mockCtrl.Finish()
	}
	return mocks.NewMockUserRepository(mockCtrl), tearDown
}

func TestCreateUser(t *testing.T) {
	type args struct {
		name     string
		email    string
		password string
	}

	tests := []struct {
		title         string
		args          args
		expectedError bool
	}{
		{
			"create user successfully",
			args{
				name:     "test name",
				email:    "test email",
				password: "test password",
			},
			false,
		},
		{
			"fail to create user",
			args{
				name:     "test name",
				email:    "test email",
				password: "test password",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mock, tearDown := getMockUserRepository(t)
			defer tearDown()
			if tt.expectedError {
				mock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(fmt.Errorf("error creating user"))
			} else {
				mock.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
			}

			usecase := usecase.NewSignupUsecase(mock, 10)
			err := usecase.Create(context.TODO(), tt.args.name, tt.args.email, tt.args.password)

			if tt.expectedError {
				require.Error(t, err, "expected an error but got none")
			} else {
				require.NoError(t, err, "expected no error but got one")
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	createdTime := time.Now()
	tests := []struct {
		title         string
		email         string
		expected      *domain.User
		expectedError bool
	}{
		{
			"get user by email successfully",
			"test email",
			&domain.User{
				ID:        1,
				Name:      "test name",
				Email:     "test email",
				Password:  "test password",
				CreatedAt: createdTime,
			},
			false,
		},
		{
			"get user by email successfully",
			"test email",
			&domain.User{
				ID:        1,
				Name:      "test name",
				Email:     "test email",
				Password:  "test password",
				CreatedAt: createdTime,
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mock, tearDown := getMockUserRepository(t)
			defer tearDown()
			if tt.expectedError {
				mock.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error getting user"))
			} else {
				mock.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).Return(tt.expected, nil)
			}

			usecase := usecase.NewSignupUsecase(mock, 10)
			user, err := usecase.GetUserByEmail(context.TODO(), tt.email)

			if tt.expectedError {
				require.Error(t, err, "expected an error but got none")
			} else {
				require.NoError(t, err, "expected no error but got one")
				require.Equal(t, tt.expected, user)
			}
		})
	}
}
