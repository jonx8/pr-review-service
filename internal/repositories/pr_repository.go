package repositories

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	m "github.com/jonx8/pr-review-service/internal/models"
)

type PRRepository interface {
	ExistsByID(prID string) (bool, error)
	GetByID(prID string) (*m.PullRequest, error)
	Create(pr *m.PullRequest) error
	UpdateStatus(prID string, status string, mergedAt *time.Time) error
	UpdateReviewer(prID string, oldUserID string, newUserID string) error
	GetByReviewer(userID string) ([]m.PullRequestShort, error)
}

type prRepository struct {
	db *sqlx.DB
}

func NewPRRepository(db *sqlx.DB) PRRepository {
	return &prRepository{db: db}
}

func (r *prRepository) ExistsByID(prID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)`
	var exists bool
	err := r.db.Get(&exists, query, prID)
	return exists, err
}

func (r *prRepository) GetByID(prID string) (*m.PullRequest, error) {
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
	err := r.db.Get(&pr, query, prID)
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
	err = r.db.Select(&reviewers, reviewersQuery, prID)
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
func (r *prRepository) Create(pr *m.PullRequest) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
        INSERT INTO pull_requests (id, title, author_id, status, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err = tx.Exec(query,
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
            INSERT INTO pr_reviewers (pr_id, user_id)
            VALUES ($1, $2)
        `
		for _, reviewerID := range pr.AssignedReviewers {
			_, err = tx.Exec(reviewerQuery, pr.PullRequestID, reviewerID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *prRepository) UpdateStatus(prID string, status string, mergedAt *time.Time) error {
	query := `
        UPDATE pull_requests 
        SET status = $1, merged_at = $2
        WHERE id = $3
    `
	_, err := r.db.Exec(query, status, mergedAt, prID)
	return err
}

func (r *prRepository) UpdateReviewer(prID string, oldUserID string, newUserID string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"DELETE FROM pr_reviewers WHERE pr_id = $1 AND user_id = $2",
		prID, oldUserID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT INTO pr_reviewers (pr_id, user_id) VALUES ($1, $2)",
		prID, newUserID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *prRepository) GetByReviewer(userID string) ([]m.PullRequestShort, error) {
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
	if err := r.db.Select(&prs, query, userID); err != nil {
		return nil, err
	}

	if prs == nil {
		return []m.PullRequestShort{}, nil
	}

	return prs, nil
}
