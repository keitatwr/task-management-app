package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/internal/session"
	"github.com/keitatwr/task-management-app/tests/helper"
	"github.com/keitatwr/task-management-app/tests/mock"
	"github.com/keitatwr/task-management-app/usecase"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type mockSessionManager struct {
	session.SessionManager
}

func (m *mockSessionManager) CreateSession(ctx *gin.Context, user domain.User) error {
	return nil
}

type errMockSessionManager struct {
	session.SessionManager
}

func (em *errMockSessionManager) CreateSession(ctx *gin.Context, user domain.User) error {
	return fmt.Errorf("failed to create session")
}

func TestFetchUserByEmail(t *testing.T) {
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
				repo.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(&domain.User{
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
			"fetch user failed",
			args{
				ctx:   context.TODO(),
				email: "test@example.com",
			},
			func(repo *mock.MockUserRepository) {
				repo.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(nil, myerror.ErrQueryFailed)
			},
			nil,
			myerror.ErrQueryFailed,
		},
		{
			"user not found",
			args{
				ctx:   context.TODO(),
				email: "test@example.com",
			},
			func(repo *mock.MockUserRepository) {
				repo.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(nil, myerror.ErrUserNotFound)
			},
			nil,
			myerror.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			mockUserRepository, teardown := helper.GetMockUserRepository(t)
			defer teardown()

			if tt.setupMockUserRepo != nil {
				tt.setupMockUserRepo(mockUserRepository)
			}

			lu := usecase.NewLoginUsecase(mockUserRepository, &mockSessionManager{})

			user, err := lu.FetchUserByEmail(tt.args.ctx, tt.args.email)

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

func TestCreateSession(t *testing.T) {
	type args struct {
		ctx  *gin.Context
		user domain.User
	}

	tests := []struct {
		title         string
		args          args
		expectedError error
	}{
		{
			"success",
			args{
				ctx: &gin.Context{},
				user: domain.User{
					ID:       1,
					Name:     "test name",
					Email:    "test email",
					Password: "test password",
				},
			},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			lu := usecase.NewLoginUsecase(nil, &mockSessionManager{})
			err := lu.CreateSession(tt.args.ctx, tt.args.user)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
