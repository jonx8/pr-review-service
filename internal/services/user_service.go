package services

import (
	"log/slog"

	"github.com/jonx8/pr-review-service/internal/errors"
	m "github.com/jonx8/pr-review-service/internal/models"
	repo "github.com/jonx8/pr-review-service/internal/repositories"
)

type UserService interface {
	GetUser(userID string) (*m.User, error)
	SetIsActive(request m.SetActiveRequest) (*m.User, error)
}

type userService struct {
	userRepository repo.UserRepository
}

func NewUserService(userRepository repo.UserRepository) UserService {
	return &userService{userRepository: userRepository}
}

func (service *userService) GetUser(userID string) (*m.User, error) {
	const method = "UserService.GetUser"

	user, err := service.userRepository.GetByID(userID)
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

func (service *userService) SetIsActive(request m.SetActiveRequest) (*m.User, error) {
	const method = "UserService.SetIsActive"

	exists, err := service.userRepository.ExistsByID(request.UserID)
	if err != nil {
		slog.Error("failed to check user existence",
			"method", method,
			"user_id", request.UserID,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to check user existence")
	}

	if !exists {
		slog.Error("user not found for activation",
			"method", method,
			"user_id", request.UserID,
		)
		return nil, errors.ErrUserNotFound
	}

	user, err := service.userRepository.SetIsActive(request.UserID, request.IsActive)
	if err != nil {
		slog.Error("failed to set user active status",
			"method", method,
			"user_id", request.UserID,
			"is_active", request.IsActive,
			"error", err,
		)
		return nil, errors.WrapInternal(err, "failed to set user active status")
	}

	return user, nil
}
