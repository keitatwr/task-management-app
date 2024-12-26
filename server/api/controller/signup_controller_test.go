package controller_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/api/controller"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/internal/security"
	"github.com/keitatwr/task-management-app/tests/helper"
	"github.com/keitatwr/task-management-app/tests/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type MockPasswordHasher struct{}

func (mh *MockPasswordHasher) HashPassword(password string) (string, error) {
	return "xxxxx", nil
}

type ErrMockPasswordHasher struct{}

func (emh *ErrMockPasswordHasher) HashPassword(password string) (string, error) {
	return "", fmt.Errorf("failed to hash password")
}

func getSignupUsecaseMock(t *testing.T) (*mock.MockSignupUsecase, func()) {
	ctrl := gomock.NewController(t)
	teardown := func() {
		ctrl.Finish()
	}
	return mock.NewMockSignupUsecase(ctrl), teardown
}

func TestSignupController(t *testing.T) {
	tests := []struct {
		title            string
		request          *http.Request
		setupMockUsecace func(signupUsecase *mock.MockSignupUsecase)
		passwordHasher   security.PasswordHasher
		wantStatus       int
		wantResponse     interface{}
	}{
		{
			"success",
			httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"name":"test","email":"test@example.com","password":"password"}`)),
			func(signupUsecase *mock.MockSignupUsecase) {
				signupUsecase.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(nil, myerror.ErrUserNotFound)
				signupUsecase.EXPECT().Create(gomock.Any(), "test", "test@example.com", "xxxxx").
					Return(nil)
			},
			&MockPasswordHasher{},
			http.StatusCreated,
			domain.SuccessResponse{Message: "user created"},
		},
		{
			"validation error one missing field",
			httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"nam":"test","email":"test@example.com","password":"password"}`)),
			nil,
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "missing fields: Name",
					},
				},
			},
		},
		{
			"validation error two missing field",
			httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"nam":"test","emai":"test@example.com","password":"password"}`)),
			nil,
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "missing fields: Name, Email",
					},
				},
			},
		},
		{
			"validation error type mismatch",
			httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"name":"test","email":"test@example.com","password":123}`)),
			nil,
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "missing field type: password, expect: string, actual: number",
					},
				},
			},
		},
		{
			"validation error json syntax error",
			httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"name":"test,"email":"test@example.com","password":"password"}`)),
			nil,
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "json syntax error, offset: 16",
					},
				},
			},
		},
		{
			"user already exists",
			httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"name":"test","email":"test@example.com","password":"password"}`)),
			func(signupUsecase *mock.MockSignupUsecase) {
				signupUsecase.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(&domain.User{}, nil)
			},
			nil,
			http.StatusConflict,
			domain.ErrorResponse{
				Message: "user already exists",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeUserAlreadyExists),
						Message:     myerror.ErrMessages[myerror.CodeUserAlreadyExists],
						Description: "email 'test@example.com' is already exists",
					},
				},
			},
		},
		{
			"failed to hash password",
			httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"name":"test","email":"test@example.com","password":"password"}`)),
			func(signupUsecase *mock.MockSignupUsecase) {
				signupUsecase.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(nil, myerror.ErrUserNotFound)
			},
			&ErrMockPasswordHasher{},
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to hash password",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeHashPasswordFailed),
						Message:     myerror.ErrMessages[myerror.CodeHashPasswordFailed],
						Description: "failed to hash password",
					},
				},
			},
		},
		{
			"failed to create user",
			httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(`{"name":"test","email":"test@example.com", "password":"password"}`)),
			func(signupUsecase *mock.MockSignupUsecase) {
				signupUsecase.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(nil, myerror.ErrUserNotFound)
				signupUsecase.EXPECT().Create(gomock.Any(), "test", "test@example.com", "xxxxx").
					Return(myerror.ErrQueryFailed)
			},
			&MockPasswordHasher{},
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to create user",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeQueryFailed),
						Message:     myerror.ErrMessages[myerror.CodeQueryFailed],
						Description: "failed to execute query",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			// mock
			signupUsecase, teardown := getSignupUsecaseMock(t)
			defer teardown()

			if tt.setupMockUsecace != nil {
				tt.setupMockUsecace(signupUsecase)
			}

			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)

			// request
			ctx.Request = tt.request

			// controller
			signupController := controller.SignupController{
				SignupUsecase:  signupUsecase,
				PasswordHasher: tt.passwordHasher,
			}

			// run
			r := gin.Default()
			r.POST("/signup", signupController.Signup)
			r.ServeHTTP(response, ctx.Request)

			// assert
			assert.Equal(t, tt.wantStatus, response.Code)
			helper.AssertResponse(t, tt.wantStatus, tt.wantResponse, response)
		})
	}
}
