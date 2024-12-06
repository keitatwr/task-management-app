package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/keitatwr/todo-app/api/controller"
	"github.com/keitatwr/todo-app/api/middleware"
	"github.com/keitatwr/todo-app/domain"
	"github.com/keitatwr/todo-app/tests/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getMockTodoUsecase(t *testing.T) (*mocks.MockTodoUsecase, func()) {
	ctrl := gomock.NewController(t)
	teardown := func() {
		ctrl.Finish()
	}
	return mocks.NewMockTodoUsecase(ctrl), teardown
}

func TestCreate(t *testing.T) {
	// mock
	todoUsecase, tearDown := getMockTodoUsecase(t)
	defer tearDown()

	tests := []struct {
		title           string
		request         domain.Todo
		expectedStatus  int
		expectedMessage string
		expectedError   bool
		invalidRequest  bool
		unauthorized    bool
		createTodoError bool
	}{
		{
			title: "success",
			request: domain.Todo{
				Title:       "title",
				Description: "description",
			},
			expectedStatus:  http.StatusCreated,
			expectedMessage: "created",
		},
		{
			title: "unsuccessfully invalid request",
			request: domain.Todo{
				Description: "description",
			},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Key: 'Todo.Title' Error:Field validation for 'Title' failed on the 'required' tag",
			expectedError:   true,
			invalidRequest:  true,
		},
		{
			title: "unsuccessfully unauthorized",
			request: domain.Todo{
				Title:       "title",
				Description: "description",
			},
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "unauthorized",
			expectedError:   true,
			unauthorized:    true,
		},
		{
			title: "unsuccessfully create todo error",
			request: domain.Todo{
				Title:       "title",
				Description: "description",
			},
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "failed to create todo",
			expectedError:   true,
			createTodoError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// request
			ctx.Request = httptest.NewRequest("POST", "/todo", strings.NewReader(
				fmt.Sprintf(`{"title":"%s","description":"%s"}`,
					tt.request.Title, tt.request.Description)))

			if !tt.unauthorized {
				user := domain.User{ID: 1, Name: "test user"}
				middleware.SetUserContext(ctx, user)
			}

			// mock expectations
			if !tt.invalidRequest && !tt.unauthorized {
				if tt.createTodoError {
					todoUsecase.EXPECT().Create(gomock.Any(), tt.request.Title, tt.request.Description, 1).
						Return(fmt.Errorf("failed to create todo"))
				} else {
					todoUsecase.EXPECT().Create(gomock.Any(), tt.request.Title, tt.request.Description, 1).
						Return(nil)
				}
			}

			// controller
			todoCotroller := controller.TodoController{TodoUsecase: todoUsecase}

			// run
			r := gin.Default()
			r.POST("/todo", todoCotroller.Create)
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

func TestGetAllTodoByUserID(t *testing.T) {
	now := time.Now()
	// mock
	todoUsecase, tearDown := getMockTodoUsecase(t)
	defer tearDown()

	tests := []struct {
		title           string
		expected        []domain.Todo
		expectedStatus  int
		expectedMessage string
		expectedError   bool
		unauthorized    bool
		getTodoError    bool
	}{
		{
			title: "success",
			expected: []domain.Todo{
				{ID: 1, Title: "title1", Description: "description1", UserID: 1, CreatedAt: now, UpdatedAt: now},
				{ID: 2, Title: "title2", Description: "description2", UserID: 1, CreatedAt: now, UpdatedAt: now},
			},
			expectedStatus: http.StatusOK,
		},
		{
			title:           "unsuccessfully unauthorized",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "unauthorized",
			expectedError:   true,
			unauthorized:    true,
		},
		{
			title:           "unsuccessfully get todo error",
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "failed to get todo",
			expectedError:   true,
			getTodoError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// request
			ctx.Request = httptest.NewRequest("GET", "/todo", nil)
			if !tt.unauthorized {
				user := domain.User{ID: 1, Name: "test user"}
				middleware.SetUserContext(ctx, user)
			}

			// mock expectations
			if !tt.unauthorized {
				if tt.getTodoError {
					todoUsecase.EXPECT().GetAllTodoByUserID(gomock.Any(), 1).
						Return(nil, fmt.Errorf("failed to get todo"))
				} else {
					todoUsecase.EXPECT().GetAllTodoByUserID(gomock.Any(), 1).
						Return(tt.expected, nil)
				}
			}

			// controller
			todoCotroller := controller.TodoController{TodoUsecase: todoUsecase}

			// run
			r := gin.Default()
			r.GET("/todo", todoCotroller.GetAllTodoByUserID)
			r.ServeHTTP(w, ctx.Request)

			// assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError {
				var response domain.ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedMessage, response.Message)
			} else {
				var response []domain.Todo
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				require.Equal(t, len(tt.expected), len(response))
				// ignore time fields
				for i := range tt.expected {
					assert.Equal(t, tt.expected[i].ID, response[i].ID)
					assert.Equal(t, tt.expected[i].Title, response[i].Title)
					assert.Equal(t, tt.expected[i].Description, response[i].Description)
					assert.Equal(t, tt.expected[i].Completed, response[i].Completed)
					assert.Equal(t, tt.expected[i].UserID, response[i].UserID)
				}
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	// mock
	todoUsecase, tearDown := getMockTodoUsecase(t)
	defer tearDown()

	tests := []struct {
		title           string
		todoID          string
		expectedStatus  int
		expectedMessage string
		expectedError   bool
		invalidRequest  bool
		unauthorized    bool
		updateTodoError bool
	}{
		{
			title:           "success",
			todoID:          "1",
			expectedStatus:  http.StatusOK,
			expectedMessage: "updated",
		},
		{
			title:           "unsuccessfully invalid request",
			todoID:          "",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "id is required",
			expectedError:   true,
			invalidRequest:  true,
		},
		{
			title:           "unsuccessfully update todo error",
			todoID:          "1",
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "failed to update todo",
			expectedError:   true,
			updateTodoError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// request
			ctx.Request = httptest.NewRequest("PUT", fmt.Sprintf("/todo?id=%v", tt.todoID), nil)

			// mock expectations
			if !tt.invalidRequest {
				if tt.updateTodoError {
					todoUsecase.EXPECT().Update(gomock.Any(), 1).
						Return(fmt.Errorf("failed to update todo"))
				} else {
					todoUsecase.EXPECT().Update(gomock.Any(), 1).
						Return(nil)
				}
			}

			// controller
			todoCotroller := controller.TodoController{TodoUsecase: todoUsecase}

			// run
			r := gin.Default()
			r.PUT("/todo", todoCotroller.Update)
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

func TestDelete(t *testing.T) {
	// mock
	todoUsecase, tearDown := getMockTodoUsecase(t)
	defer tearDown()

	tests := []struct {
		title           string
		todoID          string
		expectedStatus  int
		expectedMessage string
		expectedError   bool
		invalidRequest  bool
		unauthorized    bool
		updateTodoError bool
	}{
		{
			title:           "success",
			todoID:          "1",
			expectedStatus:  http.StatusOK,
			expectedMessage: "deleted",
		},
		{
			title:           "unsuccessfully invalid request",
			todoID:          "",
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "id is required",
			expectedError:   true,
			invalidRequest:  true,
		},
		{
			title:           "unsuccessfully update todo error",
			todoID:          "1",
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "failed to delete todo",
			expectedError:   true,
			updateTodoError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// request
			ctx.Request = httptest.NewRequest("DELETE", fmt.Sprintf("/todo?id=%v", tt.todoID), nil)

			// mock expectations
			if !tt.invalidRequest {
				if tt.updateTodoError {
					todoUsecase.EXPECT().Delete(gomock.Any(), 1).
						Return(fmt.Errorf("failed to delete todo"))
				} else {
					todoUsecase.EXPECT().Delete(gomock.Any(), 1).
						Return(nil)
				}
			}

			// controller
			todoCotroller := controller.TodoController{TodoUsecase: todoUsecase}

			// run
			r := gin.Default()
			r.DELETE("/todo", todoCotroller.Delete)
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
