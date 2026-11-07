package application

import (
	"context"

	"github.com/nicograef/jotti/backend/domain/user"
	"github.com/nicograef/jotti/backend/internal/logging"
)

type userRepo interface {
	GetUser(ctx context.Context, userID int) (user.User, error)
	CreateUser(ctx context.Context, u user.User) (int, error)
	UpdateUser(ctx context.Context, u user.User) error
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
		return 0, "", fromRepositoryError(err, log, "creation", "username", user.Username)
	}

	log.Info("User created successfully", "username", user.Username)
	return userID, onetimePassword, nil
}

// UpdateUser changes a user on behalf of actorID. Admins manage users, so another role on the
// actor's own account is always a demotion that locks the instance out.
func (c Command) UpdateUser(ctx context.Context, actorID, userID int, name, username string, role user.Role) error {
	log := logging.FromContext(ctx)

	if userID == actorID && role != user.AdminRole {
		log.Warn("Refused to demote own account", "user_id", userID)
		return ErrCannotDemoteSelf
	}

	user, err := c.UserRepo.GetUser(ctx, userID)
	if err != nil {
		return fromRepositoryError(err, log, "update", "user_id", userID)
	}

	err = user.UpdateDetails(name, username, role)
	if err != nil {
		log.Warn("Invalid user data for update", "error", err, "user_id", userID)
		return ErrInvalidUserData
	}

	err = c.UserRepo.UpdateUser(ctx, user)
	if err != nil {
		return fromRepositoryError(err, log, "update", "user_id", userID, "username", user.Username)
	}

	log.Info("User updated successfully", "user_id", userID)
	return nil
}

func (c Command) ActivateUser(ctx context.Context, userID int) error {
	return c.applyUserStatusChange(ctx, userID, "User activated successfully", func(u *user.User) { u.Activate() })
}

// DeactivateUser keeps the actor's own account active, or the last admin locks itself out.
func (c Command) DeactivateUser(ctx context.Context, actorID, userID int) error {
	if userID == actorID {
		logging.FromContext(ctx).Warn("Refused to deactivate own account", "user_id", userID)
		return ErrCannotDeactivateSelf
	}
	return c.applyUserStatusChange(ctx, userID, "User deactivated successfully", func(u *user.User) { u.Deactivate() })
}

func (c Command) DeleteUser(ctx context.Context, actorID, userID int) error {
	if userID == actorID {
		logging.FromContext(ctx).Warn("Refused to delete own account", "user_id", userID)
		return ErrCannotDeleteSelf
	}
	return c.applyUserStatusChange(ctx, userID, "User deleted successfully", func(u *user.User) { u.Delete() })
}

func (c Command) applyUserStatusChange(ctx context.Context, userID int, successMsg string, action func(*user.User)) error {
	log := logging.FromContext(ctx)

	user, err := c.UserRepo.GetUser(ctx, userID)
	if err != nil {
		return fromRepositoryError(err, log, "status change", "user_id", userID)
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
		return "", fromRepositoryError(err, log, "password reset", "user_id", userID)
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
