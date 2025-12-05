package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jonx8/pr-review-service/internal/models"
	"github.com/jonx8/pr-review-service/internal/services"
)

type PRHandler struct {
	prService services.PRService
}

func NewPRHandler(prService services.PRService) *PRHandler {
	return &PRHandler{
		prService: prService,
	}
}

func (h *PRHandler) CreatePR(c *gin.Context) {
	var req models.CreatePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, "Invalid request body: "+err.Error())
		return
	}

	pr, err := h.prService.CreatePR(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, pr)
}

func (h *PRHandler) MergePR(c *gin.Context) {
	var req models.MergePRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, "Invalid request body: "+err.Error())
		return
	}

	pr, err := h.prService.MergePR(c.Request.Context(), req.PullRequestID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, pr)
}

func (h *PRHandler) ReassignReviewer(c *gin.Context) {
	var req models.ReassignReviewerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validationError(c, "Invalid request body: "+err.Error())
		return
	}

	pr, newRevieverId, err := h.prService.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldReviewerID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pr":          pr,
		"replaced_by": newRevieverId,
	})
}

func (h *PRHandler) GetUserReviewPRs(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		validationError(c, "user_id parameter is required")
		return
	}

	prs, err := h.prService.GetPRByReviewer(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
