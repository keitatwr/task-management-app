package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/domain"
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
		// logger.Warnf(c.Request.Context(), "invalid request payload: %v", err)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Message: "invalid request payload"})
		return
	}

	// 	// check if user already exists
	// 	_, err := sc.SignupUsecase.GetUserByEmail(c, request.Email)
	// 	if err == nil {
	// 		logger.Warnf(c.Request.Context(), "email %s already exists", request.Email)
	// 		c.JSON(http.StatusConflict, domain.ErrorResponse{
	// 			Message: fmt.Sprintf("email %s already exists", request.Email)})
	// 		return
	// 	}

	// 	// hash password
	// 	request.Password, err = sc.PasswordHasher.HashPassword(request.Password)
	// 	if err != nil {
	// 		logger.Errorf(c.Request.Context(), "failed to hash password: %v", err)
	// 		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
	// 			Message: "failed to hash password"})
	// 		return
	// 	}

	// 	// create user
	// 	err = sc.SignupUsecase.Create(c, request.Name, request.Email, request.Password)
	// 	if err != nil {
	// 		logger.Errorf(c.Request.Context(), "failed to create user: %v", err)
	// 		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
	// 			Message: fmt.Sprintf("failed to create user: %v", err)})
	// 		return
	// 	}

	// c.JSON(http.StatusCreated, domain.SuccessResponse{Message: "user created"})
}
