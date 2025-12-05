package repositories

import (
	"context"
	"database/sql"
	"log/slog"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"
	m "github.com/jonx8/pr-review-service/internal/models"
)

type UserRepository interface {
	ExistsByID(ctx context.Context, userID string) (bool, error)
	GetByID(ctx context.Context, userID string) (*m.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*m.User, error)
}

type userRepository struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{
		db:     db,
		getter: trmsqlx.DefaultCtxGetter,
	}
}

func (r *userRepository) ExistsByID(ctx context.Context, userID string) (bool, error) {
	const method = "UserRepository.ExistsByID"

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	var exists bool

	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &exists, query, userID)
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

func (r *userRepository) GetByID(ctx context.Context, userID string) (*m.User, error) {
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

	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &user, query, userID)
	if err != nil {
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

func (r *userRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*m.User, error) {
	const method = "UserRepository.SetIsActive"

	query := `
		UPDATE users 
		SET is_active = $1 
		WHERE id = $2 
		RETURNING id, name, team_name, is_active
	`

	var user m.User
	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &user, query, isActive, userID)
	if err != nil {
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
