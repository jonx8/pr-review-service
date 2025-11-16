package repositories

import (
	"database/sql"
	"log/slog"

	"github.com/jmoiron/sqlx"
	m "github.com/jonx8/pr-review-service/internal/models"
)

type userRepository struct {
	db *sqlx.DB
}

type UserRepository interface {
	ExistsByID(userID string) (bool, error)
	GetByID(userID string) (*m.User, error)
	SetIsActive(userID string, isActive bool) (*m.User, error)
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) ExistsByID(userID string) (bool, error) {
	const method = "UserRepository.ExistsByID"

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	var exists bool
	err := r.db.Get(&exists, query, userID)
	if err != nil {
		slog.Error("failed to check user existence",
			"method", method,
			"user_id", userID,
			"error", err,
		)
		return false, err
	}

	return exists, nil
}

func (r *userRepository) GetByID(userID string) (*m.User, error) {
	const method = "UserRepository.GetByID"

	query := `
		SELECT 
			id, 
			name, 
			team_name, 
			is_active
		FROM users 
		WHERE id = $1
	`
	var user m.User

	if err := r.db.Get(&user, query, userID); err != nil {
		slog.Error("failed to get user by ID",
			"method", method,
			"user_id", userID,
			"error", err,
		)
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) SetIsActive(userID string, isActive bool) (*m.User, error) {
	const method = "UserRepository.SetIsActive"

	query := `
		UPDATE users 
		SET is_active = $1 
		WHERE id = $2 
		RETURNING id, name, team_name, is_active
	`

	var user m.User
	if err := r.db.Get(&user, query, isActive, userID); err != nil {
		slog.Error("failed to set user active status",
			"method", method,
			"user_id", userID,
			"is_active", isActive,
			"error", err,
		)
		return nil, err
	}

	return &user, nil
}
