package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jonx8/pr-review-service/internal/models"
	"github.com/jonx8/pr-review-service/internal/services"
)

type TeamHandler struct {
	teamService services.TeamService
}

func NewTeamHandler(teamService services.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	var team models.Team
	if err := c.ShouldBindJSON(&team); err != nil {
		validationError(c, "Invalid request body: "+err.Error())
		return
	}

	createdTeam, err := h.teamService.CreateTeam(c.Request.Context(), &team)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, createdTeam)
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		validationError(c, "team_name parameter is required")
		return
	}

	team, err := h.teamService.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, team)
}
