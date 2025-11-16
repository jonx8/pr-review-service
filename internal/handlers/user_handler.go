package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jonx8/pr-review-service/internal/models"
	"github.com/jonx8/pr-review-service/internal/services"
)

type UserHandler struct {
	userService services.UserService
	prService   services.PRService
}

func NewUserHandler(userService services.UserService, prService services.PRService) *UserHandler {
	return &UserHandler{
		userService: userService,
		prService:   prService,
	}
}

func (h *UserHandler) SetUserActive(c *gin.Context) {
	var req models.SetActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, "Invalid request body: "+err.Error())
		return
	}

	user, err := h.userService.SetIsActive(req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetUserReviewPRs(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		validationError(c, "user_id parameter is required")
		return
	}

	prs, err := h.prService.GetPRByReviewer(userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, prs)
}
