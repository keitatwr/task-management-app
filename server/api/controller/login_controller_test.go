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

type MockPasswordComparer struct{}

func (mc *MockPasswordComparer) ComparePassword(hashedPassword, password string) error {
	return nil
}

type ErrMockPasswordComparer struct{}

func (mc *ErrMockPasswordComparer) ComparePassword(hashedPassword, password string) error {
	return fmt.Errorf("invalid password")
}

func getLoginUsecaseMock(t *testing.T) (*mock.MockLoginUsecase, func()) {
	ctrl := gomock.NewController(t)
	teardown := func() {
		ctrl.Finish()
	}
	return mock.NewMockLoginUsecase(ctrl), teardown
}

func TestLogin(t *testing.T) {
	tests := []struct {
		title            string
		request          *http.Request
		setupMockUsecace func(loginUsecase *mock.MockLoginUsecase)
		passwordComparer security.PasswordComparer
		wantStatus       int
		wantResponse     interface{}
	}{
		{
			"success",
			httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"test@example.com","password":"password"}`)),
			func(loginUsecase *mock.MockLoginUsecase) {
				loginUsecase.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(&domain.User{
						Name:     "test",
						Email:    "test@example.com",
						Password: "hashedPassword",
					}, nil)
				loginUsecase.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(nil)
			},
			&MockPasswordComparer{},
			http.StatusFound,
			nil,
		},
		{
			"validation error one missing field",
			httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"test@example.com","passwor":"passeword"}`)),
			nil,
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "missing fields: Password",
					},
				},
			},
		},
		{
			"validation error type missmatch",
			httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"test@example.com","password":123}`)),
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
			httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"test@example.com, "passwor":"passeword"}`)),
			nil,
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "json syntax error, offset: 30",
					},
				},
			},
		},
		{
			"user not found",
			httptest.NewRequest("POST", "/login", strings.NewReader(`{"email":"test@example.com","password":"password"}`)),
			func(loginUsecase *mock.MockLoginUsecase) {
				loginUsecase.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(nil, myerror.ErrUserNotFound)
			},
			nil,
			http.StatusUnauthorized,
			domain.ErrorResponse{
				Message: "failed to login",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeUserNotFound),
						Message:     myerror.ErrMessages[myerror.CodeUserNotFound],
						Description: "user not found",
					},
				},
			},
		},
		{
			"fetch user failed",
			httptest.NewRequest("POST", "/login", strings.NewReader(`{"email":"test@example.com","password":"password"}`)),
			func(loginUsecase *mock.MockLoginUsecase) {
				loginUsecase.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(nil, myerror.ErrQueryFailed)
			},
			nil,
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to login",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeQueryFailed),
						Message:     myerror.ErrMessages[myerror.CodeQueryFailed],
						Description: "failed to execute query",
					},
				},
			},
		},
		{
			"invalid passwrod",
			httptest.NewRequest("POST", "/login", strings.NewReader(`{"email":"test@example.com","password":"invalid password"}`)),
			func(loginUsecase *mock.MockLoginUsecase) {
				loginUsecase.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(&domain.User{
						Name:     "test",
						Email:    "test@example.com",
						Password: "hashedPassword",
					}, nil)
			},
			&ErrMockPasswordComparer{},
			http.StatusUnauthorized,
			domain.ErrorResponse{
				Message: "failed to login",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeInvalidPassword),
						Message:     myerror.ErrMessages[myerror.CodeInvalidPassword],
						Description: "invalid password",
					},
				},
			},
		},
		{
			"create session failed",
			httptest.NewRequest("POST", "/login", strings.NewReader(`{"email":"test@example.com","password":"password"}`)),
			func(loginUsecase *mock.MockLoginUsecase) {
				loginUsecase.EXPECT().FetchUserByEmail(gomock.Any(), "test@example.com").
					Return(&domain.User{
						Name:     "test",
						Email:    "test@example.com",
						Password: "hashedPassword",
					}, nil)
				loginUsecase.EXPECT().CreateSession(gomock.Any(), gomock.Any()).Return(fmt.Errorf("failed to create session"))
			},
			&MockPasswordComparer{},
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to login",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeCreateSessionFailed),
						Message:     myerror.ErrMessages[myerror.CodeCreateSessionFailed],
						Description: "failed to create session",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			//mock
			loginUsecase, teardown := getLoginUsecaseMock(t)
			defer teardown()

			if tt.setupMockUsecace != nil {
				tt.setupMockUsecace(loginUsecase)
			}

			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)

			// request
			ctx.Request = tt.request

			// controller
			loginController := controller.LoginController{
				LoginUsecase:      loginUsecase,
				PasswordCompareer: tt.passwordComparer,
			}

			r := gin.Default()
			r.POST("/login", loginController.Login)
			r.ServeHTTP(response, ctx.Request)

			// assert
			assert.Equal(t, tt.wantStatus, response.Code)
			if tt.wantStatus != http.StatusFound {
				helper.AssertResponse(t, tt.wantStatus, tt.wantResponse, response)
			}

		})
	}
}
