package response

import (
	"github.com/gin-gonic/gin"
	"github.com/keitatwr/task-management-app/domain"
	"github.com/keitatwr/task-management-app/internal/myerror"
)

func JSON(c *gin.Context, statusCode int, message string, tasks ...domain.Task) {
	c.JSON(statusCode,
		domain.SuccessResponse{
			Message: message,
			Tasks:   tasks,
		},
	)
	return
}

func Error(c *gin.Context, statusCode int, message string, err ...error) {
	el := len(err)
	if el == 0 {
		c.JSON(statusCode, domain.ErrorResponse{Message: message})
		return
	}

	errors := make([]domain.ErrorItem, el)
	for i, e := range err {
		if appErr, ok := e.(*myerror.AppError); ok {
			errors[i] = domain.ErrorItem{
				Message:     appErr.Message,
				Code:        int(appErr.Code),
				Description: appErr.Description,
			}
		} else {
			errors[i] = domain.ErrorItem{
				Message: e.Error(),
			}
		}

		c.JSON(statusCode, domain.ErrorResponse{
			Message: message,
			Errors:  errors,
		})
		return
	}
}
