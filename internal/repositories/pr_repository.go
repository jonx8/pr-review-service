package repositories

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"
	m "github.com/jonx8/pr-review-service/internal/models"
)

type PRRepository interface {
	ExistsByID(ctx context.Context, prID string) (bool, error)
	GetByID(ctx context.Context, prID string) (*m.PullRequest, error)
	Create(ctx context.Context, pr *m.PullRequest) error
	UpdateStatus(ctx context.Context, prID string, status string, mergedAt *time.Time) error
	UpdateReviewer(ctx context.Context, prID string, oldUserID string, newUserID string) error
	GetByReviewer(ctx context.Context, userID string) ([]m.PullRequestShort, error)
}

type prRepository struct {
	db     *sqlx.DB
	getter *trmsqlx.CtxGetter
}

func NewPRRepository(db *sqlx.DB) PRRepository {
	return &prRepository{
		db:     db,
		getter: trmsqlx.DefaultCtxGetter,
	}
}

func (r *prRepository) ExistsByID(ctx context.Context, prID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)`
	var exists bool

	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &exists, query, prID)
	return exists, err
}

func (r *prRepository) GetByID(ctx context.Context, prID string) (*m.PullRequest, error) {
	const method = "PRRepository.GetByID"

	query := `
        SELECT 
            id,
            title, 
            author_id,
            status,
            created_at,
            merged_at
        FROM pull_requests
        WHERE id = $1
    `
	var pr m.PullRequest

	err := r.getter.DefaultTrOrDB(ctx, r.db).GetContext(ctx, &pr, query, prID)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Error("PR not found",
				"method", method,
				"pr_id", prID,
			)
			return nil, nil
		}
		slog.Error("failed to get PR",
			"method", method,
			"pr_id", prID,
			"error", err,
		)
		return nil, err
	}

	reviewersQuery := `
        SELECT user_id 
        FROM pr_reviewers 
        WHERE pr_id = $1
        ORDER BY assigned_at
    `
	var reviewers []string

	err = r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &reviewers, reviewersQuery, prID)
	if err != nil {
		slog.Error("failed to get PR reviewers",
			"method", method,
			"pr_id", prID,
			"error", err,
		)
		return nil, err
	}

	if reviewers == nil {
		reviewers = []string{}
	}
	pr.AssignedReviewers = reviewers

	return &pr, nil
}

func (r *prRepository) Create(ctx context.Context, pr *m.PullRequest) error {
	db := r.getter.DefaultTrOrDB(ctx, r.db)

	query := `
        INSERT INTO pull_requests (id, title, author_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := db.ExecContext(ctx, query,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
		time.Now(),
	)
	if err != nil {
		return err
	}

	if len(pr.AssignedReviewers) > 0 {
		reviewerQuery := `
            INSERT INTO pr_reviewers (pr_id, user_id, assigned_at)
            VALUES ($1, $2, $3)
        `
		assignedAt := time.Now()

		for _, reviewerID := range pr.AssignedReviewers {
			_, err = db.ExecContext(ctx, reviewerQuery,
				pr.PullRequestID, reviewerID, assignedAt)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *prRepository) UpdateStatus(ctx context.Context, prID string, status string, mergedAt *time.Time) error {
	query := `
        UPDATE pull_requests 
        SET status = $1, merged_at = $2
        WHERE id = $3
    `

	_, err := r.getter.DefaultTrOrDB(ctx, r.db).ExecContext(ctx, query, status, mergedAt, prID)
	return err
}

func (r *prRepository) UpdateReviewer(ctx context.Context, prID string, oldUserID string, newUserID string) error {
	db := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err := db.ExecContext(ctx,
		"DELETE FROM pr_reviewers WHERE pr_id = $1 AND user_id = $2",
		prID, oldUserID,
	)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(ctx,
		"INSERT INTO pr_reviewers (pr_id, user_id, assigned_at) VALUES ($1, $2, $3)",
		prID, newUserID, time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *prRepository) GetByReviewer(ctx context.Context, userID string) ([]m.PullRequestShort, error) {
	query := `
        SELECT 
            pr.id,
            pr.title,
            pr.author_id,
            pr.status
        FROM pull_requests pr
        	JOIN pr_reviewers prr ON pr.id = prr.pr_id
        	WHERE prr.user_id = $1
        ORDER BY pr.created_at DESC
    `
	var prs []m.PullRequestShort

	err := r.getter.DefaultTrOrDB(ctx, r.db).SelectContext(ctx, &prs, query, userID)
	if err != nil {
		return nil, err
	}

	if prs == nil {
		return []m.PullRequestShort{}, nil
	}

	return prs, nil
}
