package controller_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/api/controller"
	"github.com/keitatwr/task-management-app/api/middleware"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/tests/helper"
	"github.com/keitatwr/task-management-app/tests/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var TaskUsecase domain.TaskUsecase

func getMockTaskUsecase(t *testing.T) (*mock.MockTaskUsecase, func()) {
	ctrl := gomock.NewController(t)
	teardown := func() {
		ctrl.Finish()
	}
	return mock.NewMockTaskUsecase(ctrl), teardown
}

func TestTaskCtrlCreate(t *testing.T) {
	// test cases
	tests := []struct {
		title       string
		request     *http.Request
		setupMock   func(*mock.MockTaskUsecase)
		wantStatus  int
		wantRespose interface{}
	}{
		{
			"success",
			httptest.NewRequest("POST", "/tasks",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Create(gomock.Any(), "test title", "test description", 1, gomock.Any()).
					Return(nil)
			},
			http.StatusCreated,
			domain.SuccessResponse{Message: "created"},
		},
		{
			"validation error one missing field",
			httptest.NewRequest("POST", "/tasks",
				strings.NewReader(`{"description":"test description", "dueDate":"2024-12-31"}`)),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "missing fields: Title",
					},
				},
			},
		},
		{
			"validation error two missing fields",
			httptest.NewRequest("POST", "/tasks",
				strings.NewReader(`{"dueDate":"2024-12-31"}`)),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "missing fields: Title, Description",
					},
				},
			},
		},
		{
			"validation error type mismatch",
			httptest.NewRequest("POST", "/tasks",
				strings.NewReader(`{"title":1,"description":"test description", "dueDate":"2024-12-31"}`)),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "missing field type: title, expect: string, actual: number",
					},
				},
			},
		},
		{
			"validation error json syntax error",
			httptest.NewRequest("POST", "/tasks",
				strings.NewReader(`{"title":"test title, "description":"test description", "dueDate":"2024-12-31"}`)),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "json syntax error, offset: 24",
					},
				},
			},
		},
		{
			"validation error time parse error",
			httptest.NewRequest("POST", "/tasks",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-"}`)),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "time parse error, expect format: yyyy-mm-dd",
					},
				},
			},
		},
		{
			"user not found",
			httptest.NewRequest("POST", "/tasks",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			nil,
			http.StatusUnauthorized,
			domain.ErrorResponse{
				Message: "unauthorized",
				Errors: []domain.ErrorItem{
					{
						Message:     myerror.ErrMessages[myerror.CodeContextUserNotFound],
						Code:        int(myerror.CodeContextUserNotFound),
						Description: "user not found in context",
					},
				},
			},
		},
		{
			"create task DB error",
			httptest.NewRequest("POST", "/tasks",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Create(gomock.Any(), "test title", "test description", 1, gomock.Any()).
					Return(myerror.ErrQueryFailed)
			},
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to create task",
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
			"create task failed grant permission",
			httptest.NewRequest("POST", "/tasks",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Create(gomock.Any(), "test title", "test description", 1, gomock.Any()).
					Return(myerror.ErrGrantPermission)
			},
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to create task",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeGrantPermissionFailed),
						Message:     myerror.ErrMessages[myerror.CodeGrantPermissionFailed],
						Description: "failed to grant permission",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			// mock
			taskUsecase, tearDown := getMockTaskUsecase(t)
			defer tearDown()

			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)

			// request
			ctx.Request = tt.request

			if tt.wantStatus != http.StatusUnauthorized {
				middleware.SetUserContext(ctx, domain.User{ID: 1, Name: "test user"})
			}

			if tt.setupMock != nil {
				tt.setupMock(taskUsecase)
			}

			// controller
			taskCotroller := controller.TaskController{TaskUsecase: taskUsecase}

			// run
			r := gin.Default()
			r.POST("/tasks", taskCotroller.Create)
			r.ServeHTTP(response, ctx.Request)

			// assert
			assert.Equal(t, tt.wantStatus, response.Code)
			helper.AssertResponse(t, tt.wantStatus, tt.wantRespose, response)
		})
	}
}

func TestTaskCtrlFetchAllTaskByUserID(t *testing.T) {
	// test cases
	tests := []struct {
		title       string
		setupMock   func(*mock.MockTaskUsecase)
		wantStatus  int
		wantRespose interface{}
	}{
		{
			"success",
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().FetchAllTaskByUserID(gomock.Any(), 1).
					Return([]domain.Task{
						{ID: 1, Title: "title1", Description: "description1", CreatedBy: 1, DueDate: domain.NewDateOnly("2024-12-31")},
						{ID: 2, Title: "title2", Description: "description2", CreatedBy: 1, DueDate: domain.NewDateOnly("2024-12-31")},
					}, nil)
			},
			http.StatusOK,
			domain.SuccessResponse{
				Message: "fetched",
				Tasks: []domain.Task{
					{ID: 1, Title: "title1", Description: "description1", CreatedBy: 1, DueDate: domain.NewDateOnly("2024-12-31")},
					{ID: 2, Title: "title2", Description: "description2", CreatedBy: 1, DueDate: domain.NewDateOnly("2024-12-31")},
				},
			},
		},
		{
			"user not found",
			nil,
			http.StatusUnauthorized,
			domain.ErrorResponse{
				Message: "unauthorized",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeContextUserNotFound),
						Message:     myerror.ErrMessages[myerror.CodeContextUserNotFound],
						Description: "user not found in context",
					},
				},
			},
		},
		{
			"task not found",
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().FetchAllTaskByUserID(gomock.Any(), 1).
					Return(nil, myerror.ErrTaskNotFound)
			},
			http.StatusNotFound,
			domain.ErrorResponse{
				Message: "failed to fetch task",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeTaskNotFound),
						Message:     myerror.ErrMessages[myerror.CodeTaskNotFound],
						Description: "no task yet",
					},
				},
			},
		},
		{
			"permission not found",
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().FetchAllTaskByUserID(gomock.Any(), 1).
					Return(nil, myerror.ErrPermissionNotFound)
			},
			http.StatusForbidden,
			domain.ErrorResponse{
				Message: "failed to fetch task",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodePermissionNotFound),
						Message:     myerror.ErrMessages[myerror.CodePermissionNotFound],
						Description: "you don't have permission to access task",
					},
				},
			},
		},
		{
			"query error",
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().FetchAllTaskByUserID(gomock.Any(), 1).
					Return(nil, myerror.ErrQueryFailed)
			},
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to fetch task",
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
			taskUsecase, tearDown := getMockTaskUsecase(t)
			defer tearDown()

			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)

			// request
			ctx.Request = httptest.NewRequest("GET", "/tasks", nil)

			if tt.wantStatus != http.StatusUnauthorized {
				user := domain.User{ID: 1, Name: "test user"}
				middleware.SetUserContext(ctx, user)
			}

			if tt.setupMock != nil {
				tt.setupMock(taskUsecase)
			}

			// controller
			taskCotroller := controller.TaskController{TaskUsecase: taskUsecase}

			// run
			r := gin.Default()
			r.GET("/tasks", taskCotroller.FetchAllTaskByUserID)
			r.ServeHTTP(response, ctx.Request)

			// assert
			assert.Equal(t, tt.wantStatus, response.Code)
			helper.AssertResponse(t, tt.wantStatus, tt.wantRespose, response)
		})
	}
}

func TestCtrlFetchTaskByTaskID(t *testing.T) {
	tests := []struct {
		title       string
		request     *http.Request
		setupMock   func(*mock.MockTaskUsecase)
		wantStatus  int
		wantRespose interface{}
	}{
		{
			"success",
			httptest.NewRequest("GET", "/tasks/1", nil),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().FetchTaskByTaskID(gomock.Any(), 1, 1).
					Return(&domain.Task{ID: 1, Title: "title1", Description: "description1", CreatedBy: 1, DueDate: domain.NewDateOnly("2024-12-31")}, nil)
			},
			http.StatusOK,
			domain.SuccessResponse{
				Message: "fetched",
				Tasks:   []domain.Task{{ID: 1, Title: "title1", Description: "description1", CreatedBy: 1, DueDate: domain.NewDateOnly("2024-12-31")}},
			},
		},
		{
			"validation error",
			httptest.NewRequest("GET", "/tasks/abc", nil),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "string convert error, expect format: number",
					},
				},
			},
		},
		{
			"user not found",
			httptest.NewRequest("GET", "/tasks/1", nil),
			nil,
			http.StatusUnauthorized,
			domain.ErrorResponse{
				Message: "unauthorized",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeContextUserNotFound),
						Message:     myerror.ErrMessages[myerror.CodeContextUserNotFound],
						Description: "user not found in context",
					},
				},
			},
		},
		{
			"task not found",
			httptest.NewRequest("GET", "/tasks/1", nil),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().FetchTaskByTaskID(gomock.Any(), 1, 1).
					Return(nil, myerror.ErrTaskNotFound)
			},
			http.StatusNotFound,
			domain.ErrorResponse{
				Message: "failed to fetch task",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeTaskNotFound),
						Message:     myerror.ErrMessages[myerror.CodeTaskNotFound],
						Description: "no task yet",
					},
				},
			},
		},
		{
			"permission denied",
			httptest.NewRequest("GET", "/tasks/1", nil),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().FetchTaskByTaskID(gomock.Any(), 1, 1).
					Return(nil, myerror.ErrPermissionDenied)
			},
			http.StatusForbidden,
			domain.ErrorResponse{
				Message: "failed to fetch task",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodePermissionDenied),
						Message:     myerror.ErrMessages[myerror.CodePermissionDenied],
						Description: "permission denied",
					},
				},
			},
		},
		{
			"query error",
			httptest.NewRequest("GET", "/tasks/1", nil),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().FetchTaskByTaskID(gomock.Any(), 1, 1).
					Return(nil, myerror.ErrQueryFailed)
			},
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to fetch task",
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
			// mock
			taskUsecase, tearDown := getMockTaskUsecase(t)
			defer tearDown()

			gin.SetMode(gin.TestMode)

			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)

			// request
			ctx.Request = tt.request

			// user context
			if tt.wantStatus != http.StatusUnauthorized {
				user := domain.User{ID: 1, Name: "test user"}
				middleware.SetUserContext(ctx, user)
			}

			if tt.setupMock != nil {
				tt.setupMock(taskUsecase)
			}

			// controller
			taskCotroller := controller.TaskController{TaskUsecase: taskUsecase}

			// run
			r := gin.Default()
			r.GET("/tasks/:taskID", taskCotroller.FetchTaskByTaskID)
			r.ServeHTTP(response, ctx.Request)

			// assert
			assert.Equal(t, tt.wantStatus, response.Code)
			helper.AssertResponse(t, tt.wantStatus, tt.wantRespose, response)
		})
	}
}

func TestTaskCtrlUpdate(t *testing.T) {
	// test cases
	tests := []struct {
		title       string
		request     *http.Request
		setupMock   func(*mock.MockTaskUsecase)
		wantStatus  int
		wantRespose interface{}
	}{
		{
			"success",
			httptest.NewRequest("PUT", "/tasks/1",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Update(gomock.Any(), 1, 1, "test title", "test description", domain.NewDateOnly("2024-12-31")).
					Return(nil)
			},
			http.StatusOK,
			domain.SuccessResponse{Message: "updated"},
		},
		{
			"validation error uri param",
			httptest.NewRequest("PUT", "/tasks/abc",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "string convert error, expect format: number",
					},
				},
			},
		},
		{
			"validation error type mismatch",
			httptest.NewRequest("PUT", "/tasks/1",
				strings.NewReader(`{"title":1,"description":"test description", "dueDate":"2024-12-31"}`)),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "missing field type: title, expect: string, actual: number",
					},
				},
			},
		},
		{
			"validation error json syntax error",
			httptest.NewRequest("PUT", "/tasks/1",
				strings.NewReader(`{"title":"test title, "description":"test description", "dueDate":"2024-12-31"}`)),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "json syntax error, offset: 24",
					},
				},
			},
		},
		{
			"validation error time parse error",
			httptest.NewRequest("PUT", "/tasks/1",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-"}`)),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "time parse error, expect format: yyyy-mm-dd",
					},
				},
			},
		},
		{
			"user not found",
			httptest.NewRequest("PUT", "/tasks/1",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			nil,
			http.StatusUnauthorized,
			domain.ErrorResponse{
				Message: "unauthorized",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeContextUserNotFound),
						Message:     myerror.ErrMessages[myerror.CodeContextUserNotFound],
						Description: "user not found in context",
					},
				},
			},
		},
		{
			"update task DB error",
			httptest.NewRequest("PUT", "/tasks/1",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Update(gomock.Any(), 1, 1, "test title", "test description", domain.NewDateOnly("2024-12-31")).
					Return(myerror.ErrQueryFailed)
			},
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to update task",
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
			"permission denied",
			httptest.NewRequest("PUT", "/tasks/1",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Update(gomock.Any(), 1, 1, "test title", "test description", domain.NewDateOnly("2024-12-31")).
					Return(myerror.ErrPermissionDenied)
			},
			http.StatusForbidden,
			domain.ErrorResponse{
				Message: "failed to update task",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodePermissionDenied),
						Message:     myerror.ErrMessages[myerror.CodePermissionDenied],
						Description: "permission denied",
					},
				},
			},
		},
		{
			"permission not found",
			httptest.NewRequest("PUT", "/tasks/1",
				strings.NewReader(`{"title":"test title", "description":"test description", "dueDate":"2024-12-31"}`)),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Update(gomock.Any(), 1, 1, "test title", "test description", domain.NewDateOnly("2024-12-31")).
					Return(myerror.ErrPermissionNotFound)
			},
			http.StatusForbidden,
			domain.ErrorResponse{
				Message: "failed to update task",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodePermissionNotFound),
						Message:     myerror.ErrMessages[myerror.CodePermissionNotFound],
						Description: "you don't have permission to access task",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			// mock
			taskUsecase, tearDown := getMockTaskUsecase(t)
			defer tearDown()

			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)

			// request
			ctx.Request = tt.request

			// user context
			if tt.wantStatus != http.StatusUnauthorized {
				user := domain.User{ID: 1, Name: "test user"}
				middleware.SetUserContext(ctx, user)
			}

			if tt.setupMock != nil {
				tt.setupMock(taskUsecase)
			}

			// controller
			taskCotroller := controller.TaskController{TaskUsecase: taskUsecase}

			// run
			r := gin.Default()
			r.PUT("/tasks/:taskID", taskCotroller.Update)
			r.ServeHTTP(response, ctx.Request)

			// assert
			assert.Equal(t, tt.wantStatus, response.Code)
			helper.AssertResponse(t, tt.wantStatus, tt.wantRespose, response)
		})
	}
}

func TestTaskCtrlDelete(t *testing.T) {
	// test cases
	tests := []struct {
		title       string
		request     *http.Request
		setupMock   func(*mock.MockTaskUsecase)
		wantStatus  int
		wantRespose interface{}
	}{
		{
			"success",
			httptest.NewRequest("DELETE", "/tasks/1", nil),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Delete(gomock.Any(), 1, 1).
					Return(nil)
			},
			http.StatusOK,
			domain.SuccessResponse{Message: "deleted"},
		},
		{
			"validation error uri param",
			httptest.NewRequest("DELETE", "/tasks/abc", nil),
			nil,
			http.StatusBadRequest,
			domain.ErrorResponse{
				Message: "your request is validation failed",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeValidtaionFailed),
						Message:     myerror.ErrMessages[myerror.CodeValidtaionFailed],
						Description: "string convert error, expect format: number",
					},
				},
			},
		},
		{
			"user not found",
			httptest.NewRequest("DELETE", "/tasks/1", nil),
			nil,
			http.StatusUnauthorized,
			domain.ErrorResponse{
				Message: "unauthorized",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodeContextUserNotFound),
						Message:     myerror.ErrMessages[myerror.CodeContextUserNotFound],
						Description: "user not found in context",
					},
				},
			},
		},
		{
			"delete task DB error",
			httptest.NewRequest("DELETE", "/tasks/1", nil),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Delete(gomock.Any(), 1, 1).
					Return(myerror.ErrQueryFailed)
			},
			http.StatusInternalServerError,
			domain.ErrorResponse{
				Message: "failed to delete task",
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
			"permission denied",
			httptest.NewRequest("DELETE", "/tasks/1", nil),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Delete(gomock.Any(), 1, 1).
					Return(myerror.ErrPermissionDenied)
			},
			http.StatusForbidden,
			domain.ErrorResponse{
				Message: "failed to delete task",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodePermissionDenied),
						Message:     myerror.ErrMessages[myerror.CodePermissionDenied],
						Description: "permission denied",
					},
				},
			},
		},
		{
			"permission not found",
			httptest.NewRequest("DELETE", "/tasks/1", nil),
			func(taskUsecase *mock.MockTaskUsecase) {
				taskUsecase.EXPECT().Delete(gomock.Any(), 1, 1).
					Return(myerror.ErrPermissionNotFound)
			},
			http.StatusForbidden,
			domain.ErrorResponse{
				Message: "failed to delete task",
				Errors: []domain.ErrorItem{
					{
						Code:        int(myerror.CodePermissionNotFound),
						Message:     myerror.ErrMessages[myerror.CodePermissionNotFound],
						Description: "you don't have permission to access task",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			gin.SetMode(gin.TestMode)

			// mock
			taskUsecase, tearDown := getMockTaskUsecase(t)
			defer tearDown()

			response := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(response)

			// request
			ctx.Request = tt.request

			// user context
			if tt.wantStatus != http.StatusUnauthorized {
				user := domain.User{ID: 1, Name: "test user"}
				middleware.SetUserContext(ctx, user)
			}

			if tt.setupMock != nil {
				tt.setupMock(taskUsecase)
			}

			// controller
			taskCotroller := controller.TaskController{TaskUsecase: taskUsecase}

			// run
			r := gin.Default()
			r.DELETE("/tasks/:taskID", taskCotroller.Delete)
			r.ServeHTTP(response, ctx.Request)

			// assert
			assert.Equal(t, tt.wantStatus, response.Code)
			helper.AssertResponse(t, tt.wantStatus, tt.wantRespose, response)
		})
	}
}
