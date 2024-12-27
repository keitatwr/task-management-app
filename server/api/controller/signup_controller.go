package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/keitatwr/task-management-app/api/response"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/logger"
	"github.com/keitatwr/task-management-app/internal/myerror"
	"github.com/keitatwr/task-management-app/internal/security"
)

type SignupController struct {
	SignupUsecase  domain.SignupUsecase
	PasswordHasher security.PasswordHasher
}

func (sc *SignupController) Signup(c *gin.Context) {
	// binding json request
	var request domain.SignupRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		sc.handleValidationError(c, err)
		return
	}

	// check if user already exists
	_, err := sc.SignupUsecase.FetchUserByEmail(c, request.Email)
	if err == nil {
		appErr := myerror.ErrUserAlreadyExists.WrapWithDescription(err, fmt.Sprintf("email '%s' is already exists", request.Email))
		logger.W(c.Request.Context(), "user already exsits", appErr)
		response.Error(c, http.StatusConflict, "user already exists", appErr)
		return
	}

	// hash password
	request.Password, err = sc.PasswordHasher.HashPassword(request.Password)
	if err != nil {
		appErr := myerror.ErrHashPassword.WrapWithDescription(err, "failed to hash password")
		logger.W(c.Request.Context(), "failed to hash password", appErr)
		response.Error(c, http.StatusInternalServerError, "failed to hash password", appErr)
		return
	}

	// create user
	err = sc.SignupUsecase.Create(c, request.Name, request.Email, request.Password)
	if err != nil {
		sc.handleCreateError(c, err)
		return
	}

	c.JSON(http.StatusCreated, domain.SuccessResponse{Message: "user created"})
}

func (sc *SignupController) handleValidationError(c *gin.Context, err error) {
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

	default:
		vErr = myerror.ErrUnExpected.WithDescription(err.Error())
	}

	if vErr != nil {
		logger.W(c.Request.Context(), "occurred validation error", vErr)
		response.Error(c, http.StatusBadRequest, "your request is validation failed", vErr)
	}

}

func (sc *SignupController) handleCreateError(c *gin.Context, err error) {
	ctx := c.Request.Context()

	var appErr *myerror.AppError
	if errors.As(err, &appErr) {
		switch {
		case errors.Is(appErr, myerror.ErrQueryFailed):
			err := appErr.WithDescription("failed to execute query")
			logger.E(ctx, "occurred create user eroror", err)
			response.Error(c, http.StatusInternalServerError, "failed to create user", err)
			return
		default:
			logger.E(ctx, "failed to create user", err)
			response.Error(c, http.StatusInternalServerError, "failed to create user", err)
		}
	}
}
