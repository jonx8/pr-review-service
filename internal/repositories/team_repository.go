package repositories

import (
	"context"
	"log/slog"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"
	m "github.com/jonx8/pr-review-service/internal/models"
)

type TeamRepository interface {
	ExistsByName(ctx context.Context, name string) (bool, error)
	GetTeamByName(ctx context.Context, name string) (*m.Team, error)
	CreateTeam(ctx context.Context, team *m.Team) error
}

type teamRepository struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewTeamRepository(db *sqlx.DB) TeamRepository {
	return &teamRepository{
		db:     db,
		getter: trmsqlx.DefaultCtxGetter,
	}
}

func (r *teamRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	const method = "TeamRepository.ExistsByName"

	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)`
	var exists bool

	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &exists, query, name)
	if err != nil {
		slog.Error("failed to check team existence",
			"method", method,
			"team_name", name,
			"error", err,
		)
		return false, err
	}

	return exists, nil
}

func (r *teamRepository) GetTeamByName(ctx context.Context, name string) (*m.Team, error) {
	const method = "TeamRepository.GetTeamByName"

	exists, err := r.ExistsByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if !exists {
		slog.Warn("team not found",
			"method", method,
			"team_name", name,
		)
		return nil, nil
	}

	query := `
		SELECT id, name, is_active 
		FROM users 
		WHERE team_name = $1 
		ORDER BY name
	`
	var members []m.TeamMember

	err = r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &members, query, name)
	if err != nil {
		slog.Error("failed to get team members",
			"method", method,
			"team_name", name,
			"error", err,
		)
		return nil, err
	}

	return &m.Team{
		TeamName: name,
		Members:  members,
	}, nil
}

func (r *teamRepository) CreateTeam(ctx context.Context, team *m.Team) error {
	const method = "TeamRepository.CreateTeam"

	db := r.getter.DefaultTrOrDB(ctx, r.db)

	insertTeamQuery := `INSERT INTO teams (name) VALUES ($1)`
	_, err := db.ExecContext(ctx, insertTeamQuery, team.TeamName)
	if err != nil {
		slog.Error("failed to insert team",
			"method", method,
			"team_name", team.TeamName,
			"error", err,
		)
		return err
	}

	if len(team.Members) > 0 {
		usersRows := make([]m.User, len(team.Members))

		for i, member := range team.Members {
			usersRows[i] = m.User{
				UserID:   member.UserID,
				Username: member.Username,
				TeamName: team.TeamName,
				IsActive: member.IsActive,
			}
		}

		insertUsersBatchQuery := `
			INSERT INTO users (id, name, team_name, is_active)
			VALUES (:id, :name, :team_name, :is_active)
			ON CONFLICT (id) 
			DO UPDATE SET
				name = EXCLUDED.name,
				team_name = EXCLUDED.team_name,
				is_active = EXCLUDED.is_active
		`

		_, err = sqlx.NamedExecContext(ctx, db, insertUsersBatchQuery, usersRows)
		if err != nil {
			slog.Error("failed to insert team members",
				"method", method,
				"team_name", team.TeamName,
				"members_count", len(team.Members),
				"error", err,
			)
			return err
		}
	}

	return nil
}
