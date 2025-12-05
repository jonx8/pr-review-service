package services

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jonx8/pr-review-service/internal/errors"
	m "github.com/jonx8/pr-review-service/internal/models"
	repo "github.com/jonx8/pr-review-service/internal/repositories"
	u "github.com/jonx8/pr-review-service/internal/utils"
)

type PRService interface {
	GetPR(ctx context.Context, prID string) (*m.PullRequest, error)
	CreatePR(ctx context.Context, request m.CreatePRRequest) (*m.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*m.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID string, oldUserID string) (resultPR *m.PullRequest, newReviewerID *string, retErr error)
	GetPRByReviewer(ctx context.Context, userID string) ([]m.PullRequestShort, error)
}

type prService struct {
	prRepo      repo.PRRepository
	userService UserService
	teamService TeamService
	trManager   *manager.Manager
}

func NewPRService(prRepo repo.PRRepository, userService UserService, teamService TeamService, trManager *manager.Manager) PRService {
	return &prService{
		prRepo:      prRepo,
		userService: userService,
		teamService: teamService,
		trManager:   trManager,
	}
}

func (s *prService) GetPR(ctx context.Context, prID string) (*m.PullRequest, error) {
	const method = "PRService.GetPR"

	pr, err := s.prRepo.GetByID(ctx, prID)
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

func (s *prService) CreatePR(ctx context.Context, request m.CreatePRRequest) (*m.PullRequest, error) {
	const method = "PRService.CreatePR"

	var createdPR *m.PullRequest
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		exists, err := s.prRepo.ExistsByID(ctx, request.PullRequestID)
		if err != nil {
			slog.Error("failed to check PR existence",
				"method", method,
				"pr_id", request.PullRequestID,
				"error", err,
			)
			return errors.WrapInternal(err, "failed to check PR existence")
		}
		if exists {
			slog.Error("PR already exists",
				"method", method,
				"pr_id", request.PullRequestID,
			)
			return errors.ErrPRExists
		}

		author, err := s.userService.GetUser(ctx, request.AuthorID)
		if err != nil || author == nil {
			return err
		}

		team, err := s.teamService.GetTeam(ctx, author.TeamName)
		if err != nil || team == nil {
			return err
		}

		reviewers := findReviewersForPR(team, author.UserID)

		pr := &m.PullRequest{
			PullRequestID:     request.PullRequestID,
			PullRequestName:   request.PullRequestName,
			AuthorID:          request.AuthorID,
			Status:            m.StatusOpen,
			AssignedReviewers: reviewers,
		}

		if err := s.prRepo.Create(ctx, pr); err != nil {
			slog.Error("failed to create PR",
				"method", method,
				"pr_id", request.PullRequestID,
				"author_id", request.AuthorID,
				"error", err,
			)
			return errors.WrapInternal(err, "failed to create PR")
		}

		createdPR = pr
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdPR, nil
}

func (s *prService) MergePR(ctx context.Context, prID string) (*m.PullRequest, error) {
	const method = "PRService.MergePR"

	var mergedPR *m.PullRequest
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		pr, err := s.GetPR(ctx, prID)
		if err != nil || pr == nil {
			return err
		}

		if pr.Status == m.StatusMerged {
			slog.Warn("PR already merged",
				"method", method,
				"pr_id", prID,
			)
			mergedPR = pr
			return nil
		}

		mergedAt := time.Now()
		if err := s.prRepo.UpdateStatus(ctx, prID, string(m.StatusMerged), &mergedAt); err != nil {
			slog.Error("failed to update PR status",
				"method", method,
				"pr_id", prID,
				"error", err,
			)
			return errors.WrapInternal(err, "failed to update PR status")
		}

		pr.Status = m.StatusMerged
		pr.MergedAt = &mergedAt
		mergedPR = pr
		return nil
	})

	if err != nil {
		return nil, err
	}

	return mergedPR, nil
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

func (s *prService) ReassignReviewer(ctx context.Context, prID string, oldUserID string) (resultPR *m.PullRequest, newReviewerID *string, retErr error) {
	const method = "PRService.ReassignReviewer"

	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		pr, err := s.GetPR(ctx, prID)
		if err != nil || pr == nil {
			return err
		}

		if pr.Status == m.StatusMerged {
			slog.Error("cannot reassign on merged PR",
				"method", method,
				"pr_id", prID,
			)
			return errors.ErrPRMerged
		}

		oldReviewer, err := s.userService.GetUser(ctx, oldUserID)
		if err != nil || oldReviewer == nil {
			return err
		}

		team, err := s.teamService.GetTeam(ctx, oldReviewer.TeamName)
		if err != nil || team == nil {
			return err
		}

		if !u.Contains(pr.AssignedReviewers, oldUserID) {
			slog.Error("reviewer is not assigned to this PR",
				"method", method,
				"pr_id", prID,
				"old_user_id", oldUserID,
			)
			return errors.ErrNotAssigned
		}

		replacement := s.findReplacementReviewer(team, pr.AuthorID, pr.AssignedReviewers, oldUserID)
		if replacement == nil {
			slog.Error("no active replacement candidate in team",
				"method", method,
				"team_name", oldReviewer.TeamName,
				"old_user_id", oldUserID,
			)
			return errors.ErrNoCandidate
		}

		if err := s.prRepo.UpdateReviewer(ctx, prID, oldUserID, *replacement); err != nil {
			slog.Error("failed to update reviewer",
				"method", method,
				"pr_id", prID,
				"old_user_id", oldUserID,
				"new_user_id", replacement,
				"error", err,
			)
			return errors.WrapInternal(err, "failed to update reviewer")
		}

		pr.AssignedReviewers = u.ReplaceInSlice(pr.AssignedReviewers, oldUserID, *replacement)
		resultPR = pr
		newReviewerID = replacement
		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return resultPR, newReviewerID, nil
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

func (s *prService) GetPRByReviewer(ctx context.Context, userID string) ([]m.PullRequestShort, error) {
	const method = "PRService.GetPRByReviewer"

	prs, err := s.prRepo.GetByReviewer(ctx, userID)
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
