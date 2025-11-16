package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	e "github.com/jonx8/pr-review-service/internal/errors"
	"github.com/jonx8/pr-review-service/internal/models"
)

func handleError(c *gin.Context, err error) {
	var appErr *e.AppError

	if errors.As(err, &appErr) {
		c.JSON(appErr.HTTPStatus, models.ErrorResponse{
			Error: models.ErrorDetail{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
		return
	}
	c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    e.CodeInternalError,
			Message: "Internal server error",
		},
	})
}

func validationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    e.CodeBadRequest,
			Message: message,
		},
	})
}
