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

type LoginController struct {
	LoginUsecase      domain.LoginUsecase
	PasswordCompareer security.PasswordComparer
}

func (lc *LoginController) Login(c *gin.Context) {
	// binding json request
	var request domain.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		lc.handleValidationError(c, err)
		return
	}

	// user validation by email
	user, err := lc.LoginUsecase.FetchUserByEmail(c, request.Email)
	if err != nil {
		ctx := c.Request.Context()
		var appErr *myerror.AppError
		if errors.As(err, &appErr) {
			switch {
			case errors.Is(appErr, myerror.ErrUserNotFound):
				err := appErr.WithDescription("user not found")
				logger.W(ctx, "occurred user not found error", err)
				response.Error(c, http.StatusUnauthorized, "failed to login", err)
			case errors.Is(appErr, myerror.ErrQueryFailed):
				err := appErr.WithDescription("failed to execute query")
				logger.E(ctx, "occurred fetch user error", err)
				response.Error(c, http.StatusInternalServerError, "failed to login", err)
			}
		}
		return
	}

	// password validation
	if err := lc.PasswordCompareer.ComparePassword(user.Password, request.Password); err != nil {
		appErr := myerror.ErrInvalidPassword.WrapWithDescription(err, "invalid password")
		logger.W(c.Request.Context(), "occuerred invalid password error", appErr)
		response.Error(c, http.StatusUnauthorized, "failed to login", appErr)
		return
	}

	// create session
	if err := lc.LoginUsecase.CreateSession(c, *user); err != nil {
		appErr := myerror.ErrCreateSession.WrapWithDescription(err, "failed to create session")
		logger.E(c.Request.Context(), "occurred create session error", appErr)
		response.Error(c, http.StatusInternalServerError, "failed to login", appErr)
		return
	}

	c.Redirect(http.StatusFound, "/tasks")
}

func (lc *LoginController) handleValidationError(c *gin.Context, err error) {
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
