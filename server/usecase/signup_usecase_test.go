package usecase_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/tests/helper"
	"github.com/keitatwr/task-management-app/tests/mock"
	"github.com/keitatwr/task-management-app/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSignupUsecaseCreate(t *testing.T) {
	type args struct {
		ctx      context.Context
		name     string
		email    string
		password string
	}
	tests := []struct {
		title             string
		args              args
		setupMockUserRepo func(repo *mock.MockUserRepository)
		wantError         error
	}{
		{
			"success",
			args{
				ctx:      context.TODO(),
				name:     "test",
				email:    "test@example.com",
				password: "password",
			},
			func(repo *mock.MockUserRepository) {
				repo.EXPECT().Create(gomock.Any(), &domain.User{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				}).Return(nil)
			},
			nil,
		},
		{
			"create user failed",
			args{
				ctx:      context.TODO(),
				name:     "test",
				email:    "test@example.com",
				password: "password",
			},
			func(repo *mock.MockUserRepository) {
				repo.EXPECT().Create(gomock.Any(), &domain.User{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				}).Return(myerror.ErrQueryFailed)
			},
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockUerRepo, tearDown := helper.GetMockUserRepository(t)
			defer tearDown()

			tt.setupMockUserRepo(mockUerRepo)

			// run
			uc := usecase.NewSignupUsecase(mockUerRepo)
			err := uc.Create(tt.args.ctx, tt.args.name, tt.args.email, tt.args.password)

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

func TestSignupUsecaseFetchUserByEmail(t *testing.T) {
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		title             string
		args              args
		setupMockUserRepo func(repo *mock.MockUserRepository)
		wantUser          *domain.User
		wantError         error
	}{
		{
			"success",
			args{
				ctx:   context.TODO(),
				email: "test@example.com",
			},
			func(repo *mock.MockUserRepository) {
				repo.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").Return(&domain.User{
					Name:     "test",
					Email:    "test@example.com",
					Password: "password",
				}, nil)
			},
			&domain.User{
				Name:     "test",
				Email:    "test@example.com",
				Password: "password",
			},
			nil,
		},
		{
			"user not found",
			args{
				ctx:   context.TODO(),
				email: "test@example.com",
			},
			func(repo *mock.MockUserRepository) {
				repo.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").Return(nil, myerror.ErrUserNotFound)
			},
			nil,
			myerror.ErrUserNotFound,
		},
		{
			"fetch user failed",
			args{
				ctx:   context.TODO(),
				email: "test@example.com",
			},
			func(repo *mock.MockUserRepository) {
				repo.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").Return(nil, myerror.ErrQueryFailed)
			},
			nil,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockUerRepo, tearDown := helper.GetMockUserRepository(t)
			defer tearDown()

			tt.setupMockUserRepo(mockUerRepo)

			// run
			uc := usecase.NewSignupUsecase(mockUerRepo)
			user, err := uc.FetchUserByEmail(tt.args.ctx, tt.args.email)

			// assert
			if tt.wantError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantError, err)
			} else {
				assert.NoError(t, err)
				opt := cmpopts.IgnoreFields(domain.User{}, "CreatedAt")
				if diff := cmp.Diff(tt.wantUser, user, opt); diff != "" {
					t.Errorf("diff: (-want +got)\n%s", diff)
				}
			}
		})
	}
}
