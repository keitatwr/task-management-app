package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/api/middleware"
	"github.com/keitatwr/todo-app/domain"
	"github.com/keitatwr/todo-app/internal/logger"
)

type TodoController struct {
	TodoUsecase domain.TodoUsecase
}

func (tc *TodoController) Create(c *gin.Context) {
	// binding json request
	var request domain.Todo
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warnf(c.Request.Context(), "invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "invalid request payload"})
		return
	}

	// get user from context
	user := middleware.GetUserContext(c)
	if user == nil {
		logger.Warnf(c.Request.Context(), "unauthorized")
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}

	// create todo
	if err := tc.TodoUsecase.Create(c, request.Title, request.Description, user.ID); err != nil {
		logger.Errorf(c.Request.Context(), "failed to create todo: %v", err)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "failed to create todo"})
		return
	}

	c.JSON(http.StatusCreated, domain.SuccessResponse{Message: "created"})
}

func (tc *TodoController) GetAllTodoByUserID(c *gin.Context) {
	// get user from context
	user := middleware.GetUserContext(c)
	if user == nil {
		logger.Warnf(c.Request.Context(), "unauthorized")
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Message: "unauthorized"})
		return
	}
	// get all todo by user id
	todos, err := tc.TodoUsecase.GetAllTodoByUserID(c, user.ID)
	if err != nil {
		logger.Errorf(c.Request.Context(), "failed to get all todo: %v", err)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: "failed to get all todo"})
		return
	}
	c.JSON(http.StatusOK, todos)
}

func (tc *TodoController) Update(c *gin.Context) {
	// get id from path
	strID := c.Query("id")
	if strID == "" {
		logger.Warn(c.Request.Context(), "param 'id' is required")
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "param 'id' is required"})
		return
	}

	// id to int
	id, err := strconv.Atoi(strID)
	if err != nil {
		logger.Warnf(c.Request.Context(), "id must be integer: %v", err)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "id must be integer"})
		return
	}

	// update todo
	if err := tc.TodoUsecase.Update(c, id); err != nil {
		logger.Errorf(c.Request.Context(), "failed to update todo: %v", err)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "updated"})
}

func (tc *TodoController) Delete(c *gin.Context) {
	// get id from path
	strID := c.Query("id")
	if strID == "" {
		logger.Warn(c.Request.Context(), "param 'id' is required")
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "id is required"})
		return
	}

	// id to int
	id, err := strconv.Atoi(strID)
	if err != nil {
		logger.Warnf(c.Request.Context(), "id must be integer: %v", err)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Message: "id must be integer"})
		return
	}

	// update todo
	if err := tc.TodoUsecase.Delete(c, id); err != nil {
		logger.Errorf(c.Request.Context(), "failed to delete todo: %v", err)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "deleted"})
}
