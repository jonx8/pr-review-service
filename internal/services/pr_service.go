package services

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/jonx8/pr-review-service/internal/errors"
	m "github.com/jonx8/pr-review-service/internal/models"
	repo "github.com/jonx8/pr-review-service/internal/repositories"
	u "github.com/jonx8/pr-review-service/internal/utils"
)

type PRService interface {
	GetPR(PrID string) (*m.PullRequest, error)
	CreatePR(request m.CreatePRRequest) (*m.PullRequest, error)
	MergePR(prID string) (*m.PullRequest, error)
	ReassignReviewer(prID string, oldUserID string) (*m.PullRequest, *string, error)
	GetPRByReviewer(userID string) ([]m.PullRequestShort, error)
}

type prService struct {
	prRepo      repo.PRRepository
	userService UserService
	teamService TeamService
}

func NewPRService(prRepo repo.PRRepository, userService UserService, teamService TeamService) PRService {
	return &prService{
		prRepo:      prRepo,
		userService: userService,
		teamService: teamService,
	}
}

func (s *prService) GetPR(prID string) (*m.PullRequest, error) {
	const method = "PRService.GetPR"

	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		slog.Error("failed to get PR",
			"method", method,
			"pr_id", prID,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to get PR")
	}

	if pr == nil {
		slog.Error("PR not found",
			"method", method,
			"pr_id", prID,
		)
		return nil, errors.ErrPRNotFound
	}

	return pr, nil
}

func (s *prService) CreatePR(request m.CreatePRRequest) (*m.PullRequest, error) {
	const method = "PRService.CreatePR"

	exists, err := s.prRepo.ExistsByID(request.PullRequestID)
	if err != nil {
		slog.Error("failed to check PR existence",
			"method", method,
			"pr_id", request.PullRequestID,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to check PR existence")
	}
	if exists {
		slog.Error("PR already exists",
			"method", method,
			"pr_id", request.PullRequestID,
		)
		return nil, errors.ErrPRExists
	}

	author, err := s.userService.GetUser(request.AuthorID)
	if err != nil || author == nil {
		return nil, err
	}

	team, err := s.teamService.GetTeam(author.TeamName)
	if err != nil || team == nil {
		return nil, err
	}

	reviewers := findReviewersForPR(team, author.UserID)

	pr := &m.PullRequest{
		PullRequestID:     request.PullRequestID,
		PullRequestName:   request.PullRequestName,
		AuthorID:          request.AuthorID,
		Status:            m.StatusOpen,
		AssignedReviewers: reviewers,
	}

	if err := s.prRepo.Create(pr); err != nil {
		slog.Error("failed to create PR",
			"method", method,
			"pr_id", request.PullRequestID,
			"author_id", request.AuthorID,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to create PR")
	}

	return pr, nil
}

func (s *prService) MergePR(prID string) (*m.PullRequest, error) {
	const method = "PRService.MergePR"

	pr, err := s.GetPR(prID)
	if err != nil || pr == nil {
		return nil, err
	}

	if pr.Status == m.StatusMerged {
		slog.Warn("PR already merged",
			"method", method,
			"pr_id", prID,
		)
		return pr, nil
	}

	mergedAt := time.Now()
	if err := s.prRepo.UpdateStatus(prID, string(m.StatusMerged), &mergedAt); err != nil {
		slog.Error("failed to update PR status",
			"method", method,
			"pr_id", prID,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to update PR status")
	}

	pr.Status = m.StatusMerged
	pr.MergedAt = &mergedAt
	return pr, nil
}

func findReviewersForPR(team *m.Team, authorID string) []string {
	var candidates []string
	for _, member := range team.Members {
		if member.UserID != authorID && member.IsActive {
			candidates = append(candidates, member.UserID)
		}
	}

	if len(candidates) == 0 {
		return []string{}
	}

	if len(candidates) <= 2 {
		return candidates
	}

	// Generate a permutation of indices
	perm := rand.Perm(len(candidates))
	selected := []string{
		candidates[perm[0]],
		candidates[perm[1]],
	}

	return selected
}

func (s *prService) ReassignReviewer(prID string, oldUserID string) (*m.PullRequest, *string, error) {
	const method = "PRService.ReassignReviewer"

	pr, err := s.GetPR(prID)
	if err != nil || pr == nil {
		return nil, nil, err
	}

	if pr.Status == m.StatusMerged {
		slog.Error("cannot reassign on merged PR",
			"method", method,
			"pr_id", prID,
		)
		return nil, nil, errors.ErrPRMerged
	}

	oldReviewer, err := s.userService.GetUser(oldUserID)
	if err != nil || oldReviewer == nil {
		return nil, nil, err
	}

	team, err := s.teamService.GetTeam(oldReviewer.TeamName)
	if err != nil || team == nil {
		return nil, nil, err
	}

	if !u.Contains(pr.AssignedReviewers, oldUserID) {
		slog.Error("reviewer is not assigned to this PR",
			"method", method,
			"pr_id", prID,
			"old_user_id", oldUserID,
		)
		return nil, nil, errors.ErrNotAssigned
	}

	newReviewerID := s.findReplacementReviewer(team, pr.AuthorID, pr.AssignedReviewers, oldUserID)
	if newReviewerID == nil {
		slog.Error("no active replacement candidate in team",
			"method", method,
			"team_name", oldReviewer.TeamName,
			"old_user_id", oldUserID,
		)
		return nil, nil, errors.ErrNoCandidate
	}

	if err := s.prRepo.UpdateReviewer(prID, oldUserID, *newReviewerID); err != nil {
		slog.Error("failed to update reviewer",
			"method", method,
			"pr_id", prID,
			"old_user_id", oldUserID,
			"new_user_id", newReviewerID,
			"error", err,
		)
		return nil, nil, errors.WrapInternal(err, "failed to update reviewer")
	}

	pr.AssignedReviewers = u.ReplaceInSlice(pr.AssignedReviewers, oldUserID, *newReviewerID)

	return pr, newReviewerID, nil
}

func (s *prService) findReplacementReviewer(team *m.Team, authorID string, currentReviewers []string, oldUserID string) *string {
	var candidates []string

	for _, member := range team.Members {
		if member.UserID != authorID &&
			member.IsActive &&
			member.UserID != oldUserID &&
			!u.Contains(currentReviewers, member.UserID) {
			candidates = append(candidates, member.UserID)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Return random candidate
	randomIndex := rand.Intn(len(candidates))
	return &candidates[randomIndex]
}

func (s *prService) GetPRByReviewer(userID string) ([]m.PullRequestShort, error) {
	const method = "PRService.GetPRByReviewer"

	prs, err := s.prRepo.GetByReviewer(userID)
	if err != nil {
		slog.Error("failed to get PRs by reviewer",
			"method", method,
			"user_id", userID,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to get PRs by reviewer")
	}

	return prs, nil
}
