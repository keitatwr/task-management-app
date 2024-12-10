package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/todo-app/domain"
	"github.com/keitatwr/todo-app/internal/logger"
	"github.com/keitatwr/todo-app/internal/security"
)

type LoginController struct {
	LoginUsecase      domain.LoginUsecase
	PasswordCompareer security.PasswordComparer
}

func (lc *LoginController) Login(c *gin.Context) {
	// binding json request
	var request domain.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warnf(c.Request.Context(), "invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Message: "invalid request payload"})
		return
	}

	// user validation by email
	user, err := lc.LoginUsecase.GetUserByEmail(c, request.Email)
	if err != nil {
		logger.Warnf(c.Request.Context(), "user not found: %v", err)
		c.JSON(http.StatusNotFound, domain.ErrorResponse{
			Message: "user not found"})
		return
	}

	// password validation
	if err := lc.PasswordCompareer.ComparePassword(user.Password, request.Password); err != nil {
		logger.Warnf(c.Request.Context(), "password incorrect: %v", err)
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Message: "password incorrect"})
		return
	}

	// create session
	if err := lc.LoginUsecase.CreateSession(c, *user); err != nil {
		logger.Errorf(c.Request.Context(), "failed to create session: %v", err)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Message: "failed to create session"})
		return
	}

	c.JSON(http.StatusOK, domain.SuccessResponse{Message: "login success"})
}
