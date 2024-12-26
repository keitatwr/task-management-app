package usecase_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/tests/mock"
	"github.com/keitatwr/task-management-app/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func getMockUserRepository(t *testing.T) (*mock.MockUserRepository, func()) {
	ctrl := gomock.NewController(t)
	teardown := func() {
		ctrl.Finish()
	}
	return mock.NewMockUserRepository(ctrl), teardown
}

func TestSignupUsecaseCreate(t *testing.T) {
	type args struct {
		name              string
		email             string
		password          string
		setupMockUserRepo func(repo *mock.MockUserRepository)
	}
	tests := []struct {
		title     string
		args      args
		wantError error
	}{
		{
			title: "success",
			args: args{
				name:     "test",
				email:    "test@example.com",
				password: "password",
				setupMockUserRepo: func(repo *mock.MockUserRepository) {
					repo.EXPECT().Create(gomock.Any(), &domain.User{
						Name:     "test",
						Email:    "test@example.com",
						Password: "password",
					}).Return(nil)
				},
			},
			wantError: nil,
		},
		{
			title: "create user failed",
			args: args{
				name:     "test",
				email:    "test@example.com",
				password: "password",
				setupMockUserRepo: func(repo *mock.MockUserRepository) {
					repo.EXPECT().Create(gomock.Any(), &domain.User{
						Name:     "test",
						Email:    "test@example.com",
						Password: "password",
					}).Return(myerror.ErrQueryFailed)
				},
			},
			wantError: myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockUerRepo, tearDown := getMockUserRepository(t)
			defer tearDown()

			tt.args.setupMockUserRepo(mockUerRepo)

			// run
			uc := usecase.NewSignupUsecase(mockUerRepo)
			err := uc.Create(context.Background(), tt.args.name, tt.args.email, tt.args.password)

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
		email             string
		setupMockUserRepo func(repo *mock.MockUserRepository)
	}
	tests := []struct {
		title     string
		args      args
		wantUser  *domain.User
		wantError error
	}{
		{
			"success",
			args{
				email: "test@example.com",
				setupMockUserRepo: func(repo *mock.MockUserRepository) {
					repo.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").Return(&domain.User{
						Name:     "test",
						Email:    "test@example.com",
						Password: "password",
					}, nil)
				},
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
				email: "test@example.com",
				setupMockUserRepo: func(repo *mock.MockUserRepository) {
					repo.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").Return(nil, myerror.ErrUserNotFound)
				},
			},
			nil,
			myerror.ErrUserNotFound,
		},
		{
			"fetch user failed",
			args{
				email: "test@example.com",
				setupMockUserRepo: func(repo *mock.MockUserRepository) {
					repo.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").Return(nil, myerror.ErrQueryFailed)
				},
			},
			nil,
			myerror.ErrQueryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockUerRepo, tearDown := getMockUserRepository(t)
			defer tearDown()

			tt.args.setupMockUserRepo(mockUerRepo)

			// run
			uc := usecase.NewSignupUsecase(mockUerRepo)
			user, err := uc.FetchUserByEmail(context.Background(), tt.args.email)

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
