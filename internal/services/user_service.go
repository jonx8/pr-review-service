package services

import (
	"context"
	"log/slog"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jonx8/pr-review-service/internal/errors"
	m "github.com/jonx8/pr-review-service/internal/models"
	repo "github.com/jonx8/pr-review-service/internal/repositories"
)

type UserService interface {
	GetUser(ctx context.Context, userID string) (*m.User, error)
	SetIsActive(ctx context.Context, request m.SetActiveRequest) (*m.User, error)
}

type userService struct {
	userRepository repo.UserRepository
	trManager      *manager.Manager
}

func NewUserService(userRepository repo.UserRepository, trManager *manager.Manager) UserService {
	return &userService{
		userRepository: userRepository,
		trManager:      trManager,
	}
}

func (service *userService) GetUser(ctx context.Context, userID string) (*m.User, error) {
	const method = "UserService.GetUser"

	user, err := service.userRepository.GetByID(ctx, userID)
	if err != nil {
		slog.Error("failed to get user",
			"method", method,
			"user_id", userID,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to get user")
	}

	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	return user, nil
}

func (service *userService) SetIsActive(ctx context.Context, request m.SetActiveRequest) (*m.User, error) {
	const method = "UserService.SetIsActive"

	var resultUser *m.User
	err := service.trManager.Do(ctx, func(ctx context.Context) error {
		exists, err := service.userRepository.ExistsByID(ctx, request.UserID)
		if err != nil {
			slog.Error("failed to check user existence",
				"method", method,
				"user_id", request.UserID,
				"error", err,
			)
			return errors.WrapInternal(err, "failed to check user existence")
		}

		if !exists {
			slog.Error("user not found for activation",
				"method", method,
				"user_id", request.UserID,
			)
			return errors.ErrUserNotFound
		}

		user, err := service.userRepository.SetIsActive(ctx, request.UserID, request.IsActive)
		if err != nil {
			slog.Error("failed to set user active status",
				"method", method,
				"user_id", request.UserID,
				"is_active", request.IsActive,
				"error", err,
			)
			return errors.WrapInternal(err, "failed to set user active status")
		}

		resultUser = user
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resultUser, nil
}
