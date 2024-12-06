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
	"github.com/keitatwr/todo-app/tests/mocks"
	"github.com/stretchr/testify/assert"
)

type MockPasswordHasher struct{}

func (mh *MockPasswordHasher) HashPassword(password string) (string, error) {
	return "$2a$10$wH8K9f8K9f8K9f8K9f8K9u", nil
}

type ErrMockPasswordHasher struct{}

func (emh *ErrMockPasswordHasher) HashPassword(password string) (string, error) {
	return "", fmt.Errorf("failed to hash password")
}

func getSignupUsecaseMock(t *testing.T) (*mocks.MockSignupUsecase, func()) {
	ctrl := gomock.NewController(t)
	teardown := func() {
		ctrl.Finish()
	}
	return mocks.NewMockSignupUsecase(ctrl), teardown
}

func mockGetUserByEmailForSignup(signupUsecase *mocks.MockSignupUsecase,
	tt struct {
		title           string
		request         domain.SignupRequest
		expectedStatus  int
		expectedMessage string
		expectedError   bool
		invalidRequest  bool
		hashError       bool
		isAlreadyExists bool
		createUserError bool
	}) {
	if tt.isAlreadyExists {
		signupUsecase.EXPECT().GetUserByEmail(gomock.Any(), tt.request.Email).
			Return(&domain.User{}, nil)
	} else {
		signupUsecase.EXPECT().GetUserByEmail(gomock.Any(), tt.request.Email).
			Return(nil, fmt.Errorf("user not found"))
	}
}

func mockCreateUserForSignup(signupUsecase *mocks.MockSignupUsecase,
	tt struct {
		title           string
		request         domain.SignupRequest
		expectedStatus  int
		expectedMessage string
		expectedError   bool
		invalidRequest  bool
		hashError       bool
		isAlreadyExists bool
		createUserError bool
	}) {
	hashedPassword := "$2a$10$wH8K9f8K9f8K9f8K9f8K9u"
	if tt.createUserError {
		signupUsecase.EXPECT().Create(gomock.Any(), tt.request.Name, tt.request.Email, hashedPassword).
			Return(fmt.Errorf("error creating user"))
	} else {
		signupUsecase.EXPECT().Create(gomock.Any(), tt.request.Name, tt.request.Email, hashedPassword).
			Return(nil)
	}
}

func TestSignupController(t *testing.T) {
	tests := []struct {
		title           string
		request         domain.SignupRequest
		expectedStatus  int
		expectedMessage string
		expectedError   bool
		invalidRequest  bool
		hashError       bool
		isAlreadyExists bool
		createUserError bool
	}{
		{
			title: "success",
			request: domain.SignupRequest{
				Name: "test name", Email: "test@test.co.jp", Password: "secret"},
			expectedStatus:  http.StatusCreated,
			expectedMessage: "user created",
		},
		{
			title:          "unsuccessfully invalid request",
			request:        domain.SignupRequest{Name: "test name", Email: "test@test.co.jp"},
			expectedStatus: http.StatusBadRequest,
			expectedMessage: "Key: 'SignupRequest.Password' Error:" +
				"Field validation for 'Password' failed on the 'required' tag",
			expectedError:  true,
			invalidRequest: true,
		},
		{
			title: "unsuccessfully hash error",
			request: domain.SignupRequest{Name: "test name", Email: "test@test.co.jp",
				Password: "secret"},
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "failed to hash password",
			expectedError:   true,
			hashError:       true,
		},
		{
			title: "unsuccessfully user already exists",
			request: domain.SignupRequest{Name: "test name", Email: "test@test.co.jp",
				Password: "secret"},
			expectedStatus:  http.StatusConflict,
			expectedMessage: "user with email test@test.co.jp already exists",
			expectedError:   true,
			isAlreadyExists: true,
		},
		{
			title: "unsuccessfully create user error",
			request: domain.SignupRequest{Name: "test name", Email: "test@test.co.jp",
				Password: "secret"},
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "failed to create user: error creating user",
			expectedError:   true,
			createUserError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			// mock
			signupUsecase, tearDown := getSignupUsecaseMock(t)
			defer tearDown()
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// mock expectations
			if !tt.invalidRequest {
				mockGetUserByEmailForSignup(signupUsecase, tt)
				if !tt.hashError && !tt.isAlreadyExists {
					mockCreateUserForSignup(signupUsecase, tt)
				}
			}

			// request
			ctx.Request = httptest.NewRequest("POST", "/signup", strings.NewReader(
				fmt.Sprintf(`{"name":"%s","email":"%s","password":"%s"}`,
					tt.request.Name, tt.request.Email, tt.request.Password)))
			ctx.Request.Header.Set("Content-Type", "application/json")

			// controller
			signupController := controller.SignupController{SignupUsecase: signupUsecase}
			if tt.hashError {
				signupController.PasswordHasher = &ErrMockPasswordHasher{}
			} else {
				signupController.PasswordHasher = &MockPasswordHasher{}
			}

			// run
			r := gin.Default()
			r.POST("/signup", signupController.Signup)
			r.ServeHTTP(w, ctx.Request)

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
