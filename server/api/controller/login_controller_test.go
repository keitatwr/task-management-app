package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/keitatwr/todo-app/api/controller"
	"github.com/keitatwr/todo-app/domain"
	"github.com/keitatwr/todo-app/domain/mocks"
	"github.com/stretchr/testify/assert"
)

type MockPasswordComparer struct{}

func (mc *MockPasswordComparer) ComparePassword(hashedPassword, password string) error {
	return nil
}

type ErrMockPasswordComparer struct{}

func (emc *ErrMockPasswordComparer) ComparePassword(hashedPassword, password string) error {
	return fmt.Errorf("password is incorrect")
}

func getLoginUsecaseMock(t *testing.T) (*mocks.MockLoginUsecase, func()) {
	ctrl := gomock.NewController(t)
	teardown := func() {
		ctrl.Finish()
	}
	return mocks.NewMockLoginUsecase(ctrl), teardown
}

func mockGetUserByEmailForLogin(ctx *gin.Context, loginUsecase *mocks.MockLoginUsecase, tt struct {
	title              string
	request            domain.LoginRequest
	expectedStatus     int
	expectedMessage    string
	expectedError      bool
	invalidRequest     bool
	userNotFound       bool
	passwordIncorrect  bool
	createSessionError bool
}) {
	if tt.userNotFound {
		loginUsecase.EXPECT().GetUserByEmail(ctx, tt.request.Email).
			Return(nil, fmt.Errorf("user not found"))
	} else {
		loginUsecase.EXPECT().GetUserByEmail(ctx, tt.request.Email).
			Return(&domain.User{}, nil)
	}
}

func mockCreateSessionForLogin(ctx *gin.Context, loginUsecase *mocks.MockLoginUsecase, tt struct {
	title              string
	request            domain.LoginRequest
	expectedStatus     int
	expectedMessage    string
	expectedError      bool
	invalidRequest     bool
	userNotFound       bool
	passwordIncorrect  bool
	createSessionError bool
}) {
	if tt.createSessionError {
		loginUsecase.EXPECT().CreateSession(ctx, gomock.Any()).
			Return(fmt.Errorf("failed to create session"))
	} else {
		loginUsecase.EXPECT().CreateSession(ctx, gomock.Any()).
			Return(nil)
	}
}

func TestLoginController(t *testing.T) {
	tests := []struct {
		title              string
		request            domain.LoginRequest
		expectedStatus     int
		expectedMessage    string
		expectedError      bool
		invalidRequest     bool
		userNotFound       bool
		passwordIncorrect  bool
		createSessionError bool
	}{
		{
			title: "happy path",
			request: domain.LoginRequest{
				Email:    "test@test.co.jp",
				Password: "secret",
			},
			expectedStatus:  http.StatusOK,
			expectedMessage: "login success",
		},
		{
			title: "login unsuccessfully invalid request",
			request: domain.LoginRequest{
				Email: "test@test.co.jp",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			invalidRequest: true,
			expectedMessage: "Key: 'LoginRequest.Password' Error:" +
				"Field validation for 'Password' failed on the 'required' tag",
		},
		{
			title: "login unsuccessfully user not found",
			request: domain.LoginRequest{
				Email:    "test@test.co.jp",
				Password: "secret",
			},
			expectedStatus:  http.StatusNotFound,
			expectedError:   true,
			userNotFound:    true,
			expectedMessage: "user not found",
		},
		{
			title: "login unsuccessfully password is incorrect",
			request: domain.LoginRequest{
				Email:    "test@test.co.jp",
				Password: "secret",
			},
			expectedStatus:    http.StatusUnauthorized,
			expectedError:     true,
			passwordIncorrect: true,
			expectedMessage:   "password is incorrect",
		},
		{
			title: "login unsuccessfully create session error",
			request: domain.LoginRequest{
				Email:    "test@test.co.jp",
				Password: "secret",
			},
			expectedStatus:     http.StatusInternalServerError,
			expectedError:      true,
			createSessionError: true,
			expectedMessage:    "failed to create session",
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			// mock
			loginUsecase, tearDown := getLoginUsecaseMock(t)
			defer tearDown()
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// mock expectation
			if !tt.invalidRequest {
				mockGetUserByEmailForLogin(ctx, loginUsecase, tt)
				if !tt.userNotFound && !tt.passwordIncorrect {
					mockCreateSessionForLogin(ctx, loginUsecase, tt)
				}
			}

			// request
			ctx.Request = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(
				fmt.Sprintf(`{"email":"%s","password":"%s"}`,
					tt.request.Email, tt.request.Password)))
			ctx.Request.Header.Set("Content-Type", "application/json")

			// controller
			loginController := controller.LoginController{LoginUsecase: loginUsecase}
			if tt.passwordIncorrect {
				loginController.PasswordCompareer = &ErrMockPasswordComparer{}
			} else {
				loginController.PasswordCompareer = &MockPasswordComparer{}
			}

			// run
			loginController.Login(ctx)

			// assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError {
				var response domain.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMessage, response.Message)
			} else {
				var response domain.SuccessResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMessage, response.Message)
			}
		})
	}
}
