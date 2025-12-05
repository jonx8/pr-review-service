package services

import (
	"context"
	"log/slog"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jonx8/pr-review-service/internal/errors"
	m "github.com/jonx8/pr-review-service/internal/models"
	repo "github.com/jonx8/pr-review-service/internal/repositories"
)

type TeamService interface {
	GetTeam(ctx context.Context, name string) (*m.Team, error)
	CreateTeam(ctx context.Context, team *m.Team) (*m.Team, error)
}

type teamService struct {
	teamRepository repo.TeamRepository
	trManager      *manager.Manager
}

func NewTeamService(teamRepository repo.TeamRepository, trManager *manager.Manager) TeamService {
	return &teamService{
		teamRepository: teamRepository,
		trManager:      trManager,
	}
}

func (service *teamService) CreateTeam(ctx context.Context, team *m.Team) (*m.Team, error) {
	const method = "TeamService.CreateTeam"

	err := service.trManager.Do(ctx, func(ctx context.Context) error {
		exists, err := service.teamRepository.ExistsByName(ctx, team.TeamName)
		if err != nil {
			slog.Error("failed to check team existence",
				"method", method,
				"team_name", team.TeamName,
				"error", err,
			)
			return errors.WrapInternal(err, "failed to check team existence")
		}
		if exists {
			slog.Warn("team already exists",
				"method", method,
				"team_name", team.TeamName,
			)
			return errors.ErrTeamExists
		}

		if err := service.teamRepository.CreateTeam(ctx, team); err != nil {
			slog.Error("failed to create team",
				"method", method,
				"team_name", team.TeamName,
				"error", err,
			)
			return errors.WrapInternal(err, "failed to create team")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return team, nil
}

func (service *teamService) GetTeam(ctx context.Context, name string) (*m.Team, error) {
	const method = "TeamService.GetTeam"

	team, err := service.teamRepository.GetTeamByName(ctx, name)
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
