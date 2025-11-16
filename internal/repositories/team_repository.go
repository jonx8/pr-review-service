package repositories

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	m "github.com/jonx8/pr-review-service/internal/models"
)

type TeamRepository interface {
	ExistsByName(name string) (bool, error)
	GetTeamByName(name string) (*m.Team, error)
	CreateTeam(team *m.Team) error
}

type teamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) TeamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) ExistsByName(name string) (bool, error) {
	const method = "TeamRepository.ExistsByName"

	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)`
	var exists bool
	err := r.db.Get(&exists, query, name)
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

func (r *teamRepository) GetTeamByName(name string) (*m.Team, error) {
	const method = "TeamRepository.GetTeamByName"

	exists, err := r.ExistsByName(name)
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

	// Get team members
	query := `
		SELECT id, name, is_active 
		FROM users 
		WHERE team_name = $1 
		ORDER BY name
	`
	var members []m.TeamMember
	if err := r.db.Select(&members, query, name); err != nil {
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

func (r *teamRepository) CreateTeam(team *m.Team) error {
	const method = "TeamRepository.CreateTeam"

	tx, err := r.db.Beginx()
	if err != nil {
		slog.Error("failed to begin transaction",
			"method", method,
			"team_name", team.TeamName,
			"error", err,
		)
		return err
	}
	defer tx.Rollback()

	if _, err = tx.Exec("INSERT INTO teams (name) VALUES ($1)", team.TeamName); err != nil {
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

		if _, err = tx.NamedExec(insertUsersBatchQuery, usersRows); err != nil {
			slog.Error("failed to insert team members",
				"method", method,
				"team_name", team.TeamName,
				"members_count", len(team.Members),
				"error", err,
			)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("failed to commit transaction",
			"method", method,
			"team_name", team.TeamName,
			"error", err,
		)
		return err
	}

	return nil
}
