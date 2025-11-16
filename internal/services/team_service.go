package services

import (
	"fmt"
	"log/slog"

	"github.com/jonx8/pr-review-service/internal/errors"
	m "github.com/jonx8/pr-review-service/internal/models"
	repo "github.com/jonx8/pr-review-service/internal/repositories"
)

type TeamService interface {
	GetTeam(name string) (*m.Team, error)
	CreateTeam(*m.Team) (*m.Team, error)
}

type teamService struct {
	teamRepository repo.TeamRepository
}

func NewTeamService(teamRepository repo.TeamRepository) TeamService {
	return &teamService{teamRepository: teamRepository}
}

func (service *teamService) CreateTeam(team *m.Team) (*m.Team, error) {
	const method = "TeamService.CreateTeam"

	exists, err := service.teamRepository.ExistsByName(team.TeamName)
	if err != nil {
		slog.Error("failed to check team existence",
			"method", method,
			"team_name", team.TeamName,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to check team existence")
	}
	if exists {
		slog.Warn("team already exists",
			"method", method,
			"team_name", team.TeamName,
		)
		return nil, errors.ErrTeamExists
	}

	// Check for duplicatate users in the members list
	userIDs := make(map[string]bool)
	for i, member := range team.Members {
		if userIDs[member.UserID] {
			slog.Warn("duplicate user_id in team members",
				"method", method,
				"team_name", team.TeamName,
				"user_id", member.UserID,
				"position", i+1,
			)
			return nil, errors.NewValidation(fmt.Sprintf("duplicate user_id '%s' at position %d", member.UserID, i+1))
		}
		userIDs[member.UserID] = true
	}

	if err := service.teamRepository.CreateTeam(team); err != nil {
		slog.Error("failed to create team",
			"method", method,
			"team_name", team.TeamName,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to create team")
	}

	return team, nil
}

func (service *teamService) GetTeam(name string) (*m.Team, error) {
	const method = "TeamService.GetTeam"

	team, err := service.teamRepository.GetTeamByName(name)
	if err != nil {
		slog.Error("failed to get team",
			"method", method,
			"team_name", name,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to get team")
	}

	if team == nil {
		slog.Warn("team not found",
			"method", method,
			"team_name", name,
		)
		return nil, errors.ErrTeamNotFound
	}

	return team, nil
}
