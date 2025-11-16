package models

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id" db:"id"`
	PullRequestName   string     `json:"pull_request_name" db:"title"`
	AuthorID          string     `json:"author_id" db:"author_id"`
	Status            PRStatus   `json:"status" db:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"-" db:"created_at"`
	MergedAt          *time.Time `json:"mergedAt,omitempty" db:"merged_at"`
}

type PullRequestShort struct {
	PullRequestID   string   `json:"pull_request_id" db:"id"`
	PullRequestName string   `json:"pull_request_name" db:"title"`
	AuthorID        string   `json:"author_id" db:"author_id"`
	Status          PRStatus `json:"status" db:"status"`
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required,min=1,max=50"`
	PullRequestName string `json:"pull_request_name" binding:"required,min=1,max=255"`
	AuthorID        string `json:"author_id" binding:"required,min=1,max=50"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldReviewerID string `json:"old_reviewer_id" binding:"required"`
}
