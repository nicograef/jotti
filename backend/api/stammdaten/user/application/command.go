package application

import (
	"context"
	"errors"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type userRepo interface {
	GetUser(ctx context.Context, userID int) (user.User, error)
	CreateUser(ctx context.Context, u user.User) (int, error)
	UpdateUser(ctx context.Context, u user.User) error
	GetAllUsers(ctx context.Context) ([]user.User, error)
}

type Command struct {
	UserRepo userRepo
}

func (c Command) CreateUser(ctx context.Context, name, username string, role user.Role) (int, string, error) {
	log := logging.FromContext(ctx)

	user, onetimePassword, err := user.NewUser(name, username, role)
	if err != nil {
		log.Warn("Invalid user data", "error", err, "username", username)
		return 0, "", ErrInvalidUserData
	}

	userID, err := c.UserRepo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn("Username already exists", "error", err, "username", user.Username)
			return 0, "", ErrUsernameAlreadyExists
		}
		log.Error("Failed to create user", "error", err, "username", user.Username)
		return 0, "", ErrDatabase
	}

	log.Info("User created successfully", "username", user.Username)
	return userID, onetimePassword, nil
}

func (c Command) UpdateUser(ctx context.Context, userID int, name, username string, role user.Role) error {
	log := logging.FromContext(ctx)

	user, err := c.UserRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("User not found for update", "user_id", userID)
			return ErrUserNotFound
		}
		log.Error("Failed to retrieve user for update", "error", err, "user_id", userID)
		return ErrDatabase
	}

	err = user.UpdateDetails(name, username, role)
	if err != nil {
		log.Warn("Invalid user data for update", "error", err, "user_id", userID)
		return ErrInvalidUserData
	}

	err = c.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn("Username already exists", "error", err, "username", user.Username)
			return ErrUsernameAlreadyExists
		}
		log.Error("Failed to update user", "error", err, "user_id", userID)
		return ErrDatabase
	}

	log.Info("User updated successfully", "user_id", userID)
	return nil
}

func (c Command) ActivateUser(ctx context.Context, userID int) error {
	return c.applyUserStatusChange(ctx, userID, "User activated successfully", func(u *user.User) { u.Activate() })
}

func (c Command) DeactivateUser(ctx context.Context, userID int) error {
	return c.applyUserStatusChange(ctx, userID, "User deactivated successfully", func(u *user.User) { u.Deactivate() })
}

func (c Command) DeleteUser(ctx context.Context, userID int) error {
	return c.applyUserStatusChange(ctx, userID, "User deleted successfully", func(u *user.User) { u.Delete() })
}

func (c Command) applyUserStatusChange(ctx context.Context, userID int, successMsg string, action func(*user.User)) error {
	log := logging.FromContext(ctx)

	user, err := c.UserRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("User not found for status change", "user_id", userID)
			return ErrUserNotFound
		}
		log.Error("Failed to retrieve user for status change", "error", err, "user_id", userID)
		return ErrDatabase
	}

	action(&user)

	err = c.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		log.Error("Failed to update user", "error", err, "user_id", userID)
		return ErrDatabase
	}

	log.Info(successMsg, "user_id", userID)
	return nil
}

func (c Command) ResetPassword(ctx context.Context, userID int) (string, error) {
	log := logging.FromContext(ctx)

	user, err := c.UserRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			log.Warn("User not found for password reset", "user_id", userID)
			return "", ErrUserNotFound
		}
		log.Error("Failed to retrieve user for password reset", "error", err, "user_id", userID)
		return "", ErrDatabase
	}

	onetimePassword, err := user.ResetPassword()
	if err != nil {
		log.Error("Failed to reset password", "error", err, "user_id", userID)
		return "", err
	}

	err = c.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		log.Error("Failed to update user in persistence", "error", err, "user_id", userID)
		return "", ErrDatabase
	}

	log.Info("Password reset successfully", "user_id", userID)
	return onetimePassword, nil
}
