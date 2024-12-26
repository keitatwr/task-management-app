package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/keitatwr/task-management-app/api/middleware"
	"github.com/keitatwr/task-management-app/api/response"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/logger"
	"github.com/keitatwr/task-management-app/internal/myerror"
)

type TaskController struct {
	TaskUsecase domain.TaskUsecase
}

func (tc *TaskController) Create(c *gin.Context) {
	// binding json request
	var request domain.TaskCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		tc.handleValidationError(c, err)
		return
	}

	// get user from context
	user := middleware.GetUserContext(c)
	if user == nil {
		err := myerror.ErrContextUserNotFound.WithDescription("user not found in context")
		logger.W(c.Request.Context(), "occurred context error", err)
		response.Error(c, http.StatusUnauthorized, "unauthorized", err)
		return
	}
	// create task
	if err := tc.TaskUsecase.Create(c, request.Title, request.Description, user.ID, request.DueDate); err != nil {
		tc.handleCreateTaskError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, "created")
}

func (tc *TaskController) FetchAllTaskByUserID(c *gin.Context) {
	// get user from context
	user := middleware.GetUserContext(c)
	if user == nil {
		err := myerror.ErrContextUserNotFound.WithDescription("user not found in context")
		logger.W(c.Request.Context(), "occurred context error", err)
		response.Error(c, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	// get all task by user id
	tasks, err := tc.TaskUsecase.FetchAllTaskByUserID(c, user.ID)
	if err != nil {
		tc.handleFetchTaskError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, "fetched", tasks...)
}

func (tc *TaskController) FetchTaskByTaskID(c *gin.Context) {
	// get id from path
	var request domain.TaskFetchRequest
	if err := c.ShouldBindUri(&request); err != nil {
		tc.handleValidationError(c, err)
		return
	}

	// get user from context
	user := middleware.GetUserContext(c)
	if user == nil {
		err := myerror.ErrContextUserNotFound.WithDescription("user not found in context")
		logger.W(c.Request.Context(), "occurred context error", err)
		response.Error(c, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	task, err := tc.TaskUsecase.FetchTaskByTaskID(c, request.ID, user.ID)
	if err != nil {
		tc.handleFetchTaskError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, "fetched", *task)
}

func (tc *TaskController) Update(c *gin.Context) {
	// get id from path
	var request domain.TaskUpdateRequest
	if err := c.ShouldBindUri(&request); err != nil {
		tc.handleValidationError(c, err)
		return
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		fmt.Println(err)
		tc.handleValidationError(c, err)
		return
	}

	// get user from context
	user := middleware.GetUserContext(c)
	if user == nil {
		err := myerror.ErrContextUserNotFound.WithDescription("user not found in context")
		logger.W(c.Request.Context(), "occurred context error", err)
		response.Error(c, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	// update task
	if err := tc.TaskUsecase.Update(c, request.ID, user.ID, request.Title, request.Description, request.DueDate); err != nil {
		tc.handleUpdateTaskError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, "updated")
}

func (tc *TaskController) Delete(c *gin.Context) {
	// get id from path
	var request domain.TaskFetchRequest
	if err := c.ShouldBindUri(&request); err != nil {
		tc.handleValidationError(c, err)
		return
	}

	// get user from context
	user := middleware.GetUserContext(c)
	if user == nil {
		err := myerror.ErrContextUserNotFound.WithDescription("user not found in context")
		logger.W(c.Request.Context(), "occurred context error", err)
		response.Error(c, http.StatusUnauthorized, "unauthorized", err)
		return
	}

	// delete task
	if err := tc.TaskUsecase.Delete(c, request.ID, user.ID); err != nil {
		tc.handleDeleteTaskError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, "deleted")
}

func (tc *TaskController) handleValidationError(c *gin.Context, err error) {
	var vErr *myerror.AppError

	switch e := err.(type) {
	case validator.ValidationErrors:
		missingFields := []string{}
		for _, fieldErr := range e {
			missingFields = append(missingFields, fieldErr.Field())
		}
		vErr = myerror.ErrValidation.WrapWithDescription(e,
			fmt.Sprintf("missing fields: %v", strings.Join(missingFields, ", ")))

	case *json.UnmarshalTypeError:
		vErr = myerror.ErrValidation.WrapWithDescription(e,
			fmt.Sprintf("missing field type: %v, expect: %s, actual: %s", e.Field, e.Type, e.Value))

	case *json.SyntaxError:
		vErr = myerror.ErrValidation.WrapWithDescription(e,
			fmt.Sprintf("json syntax error, offset: %d", e.Offset))

	case *time.ParseError:
		vErr = myerror.ErrValidation.WrapWithDescription(e,
			fmt.Sprintf("time parse error, expect format: %s", "yyyy-mm-dd"))

	case *strconv.NumError:
		vErr = myerror.ErrValidation.WrapWithDescription(e,
			"string convert error, expect format: number")

	default:
		vErr = myerror.ErrUnExpected.WithDescription(err.Error())
	}

	if vErr != nil {
		logger.W(c.Request.Context(), "occurred validation error", vErr)
		response.Error(c, http.StatusBadRequest, "your request is validation failed", vErr)
	}

}

func (tc *TaskController) handleCreateTaskError(c *gin.Context, err error) {
	ctx := c.Request.Context()

	var appErr *myerror.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr, myerror.ErrQueryFailed):
			err := appErr.WithDescription("failed to execute query")
			logger.E(ctx, "occurred db error", err)
			response.Error(c, http.StatusInternalServerError, "failed to create task", err)

		case errors.Is(appErr, myerror.ErrGrantPermission):
			err := appErr.WithDescription("failed to grant permission")
			logger.E(ctx, "occurred create task error", err)
			response.Error(c, http.StatusInternalServerError, "failed to create task", err)

		default:
			logger.E(ctx, "occurred create task error", appErr)
			response.Error(c, http.StatusInternalServerError, "failed to create task", appErr)
		}
	} else {

		logger.E(ctx, "unexpected error occurred", err)
		response.Error(c, http.StatusInternalServerError, "failed to create task", err)
	}
}

func (tc *TaskController) handleFetchTaskError(c *gin.Context, err error) {
	ctx := c.Request.Context()

	var appErr *myerror.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr, myerror.ErrQueryFailed):
			err := appErr.WithDescription("failed to execute query")
			logger.E(ctx, "occurred fetch task error", err)
			response.Error(c, http.StatusInternalServerError, "failed to fetch task", err)

		case errors.Is(appErr, myerror.ErrTaskNotFound):
			err := appErr.WithDescription("no task yet")
			logger.W(ctx, "occurred fetch task error", err)
			response.Error(c, http.StatusNotFound, "failed to fetch task", err)

		case errors.Is(appErr, myerror.ErrPermissionNotFound):
			err := appErr.WithDescription("you don't have permission to access task")
			logger.W(ctx, "occurred fetch task error", err)
			response.Error(c, http.StatusForbidden, "failed to fetch task", err)

		case errors.Is(appErr, myerror.ErrPermissionDenied):
			err := appErr.WithDescription("permission denied")
			logger.W(ctx, "occurred fetch task error", err)
			response.Error(c, http.StatusForbidden, "failed to fetch task", err)

		default:
			logger.E(ctx, "occurred fetch task error", appErr)
			response.Error(c, http.StatusInternalServerError, "failed to fetch task", appErr)
		}
	} else {
		logger.E(ctx, "unexpected error occurred", err)
		response.Error(c, http.StatusInternalServerError, "failed to fetch task", err)
	}
}

func (tc *TaskController) handleUpdateTaskError(c *gin.Context, err error) {
	ctx := c.Request.Context()

	var appErr *myerror.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr, myerror.ErrQueryFailed):
			err := appErr.WithDescription("failed to execute query")
			logger.E(ctx, "occurred update task error", err)
			response.Error(c, http.StatusInternalServerError, "failed to update task", err)

		case errors.Is(appErr, myerror.ErrPermissionNotFound):
			err := appErr.WithDescription("you don't have permission to access task")
			logger.W(ctx, "occurred update task error", err)
			response.Error(c, http.StatusForbidden, "failed to update task", err)

		case errors.Is(appErr, myerror.ErrPermissionDenied):
			err := appErr.WithDescription("permission denied")
			logger.W(ctx, "occurred update task error", err)
			response.Error(c, http.StatusForbidden, "failed to update task", err)

		default:
			logger.E(ctx, "occurred update task error", appErr)
			response.Error(c, http.StatusInternalServerError, "failed to update task", appErr)
		}
	} else {
		logger.E(ctx, "unexpected error occurred", err)
		response.Error(c, http.StatusInternalServerError, "failed to update task", err)
	}
}

func (tc *TaskController) handleDeleteTaskError(c *gin.Context, err error) {
	ctx := c.Request.Context()

	var appErr *myerror.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr, myerror.ErrQueryFailed):
			err := appErr.WithDescription("failed to execute query")
			logger.E(ctx, "occurred delete task error", err)
			response.Error(c, http.StatusInternalServerError, "failed to delete task", err)

		case errors.Is(appErr, myerror.ErrPermissionNotFound):
			err := appErr.WithDescription("you don't have permission to access task")
			logger.W(ctx, "occurred delete task error", err)
			response.Error(c, http.StatusForbidden, "failed to delete task", err)

		case errors.Is(appErr, myerror.ErrPermissionDenied):
			err := appErr.WithDescription("permission denied")
			logger.W(ctx, "occurred delete task error", err)
			response.Error(c, http.StatusForbidden, "failed to delete task", err)

		default:
			logger.E(ctx, "occurred delete task error", appErr)
			response.Error(c, http.StatusInternalServerError, "failed to delete task", appErr)
		}
	} else {
		logger.E(ctx, "unexpected error occurred", err)
		response.Error(c, http.StatusInternalServerError, "failed to delete task", err)
	}
}
